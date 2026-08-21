package dto_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/router"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func r006DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), SkipDefaultTransaction: false})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	return db, mock
}

func r006Engine(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	cfg := config.Load()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return router.Setup(cfg, db, l)
}

func r006Token(t *testing.T) string {
	t.Helper()
	cfg := config.Load()
	tok, err := util.GenerateToken(2, "gardener", "user", cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

var r006SelRe = regexp.QuoteMeta("SELECT * FROM `questions` WHERE `questions`.`id` = ? ORDER BY `questions`.`id` LIMIT ?")
var r006CountRe = regexp.QuoteMeta("SELECT count(*) FROM `questions`")
var r006ListRe = regexp.QuoteMeta("SELECT * FROM `questions` ORDER BY id DESC LIMIT ?")

func TestQuestionGetNilMapNoPanic(t *testing.T) {
	db, mock := r006DB(t)
	mock.ExpectQuery(r006SelRe).WithArgs(1, 1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "title", "content", "images", "status", "created_at"}).
			AddRow(1, 2, "title", "content", "[]", "open", time.Now()))
	r := r006Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/questions/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first question visit must not panic, status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestQuestionZeroIdBad(t *testing.T) {
	db, mock := r006DB(t)
	r := r006Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/questions/0", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatalf("question id=0 must not return 200, body=%s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestQuestionCreateNoPanic(t *testing.T) {
	db, mock := r006DB(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `questions` (`user_id`,`title`,`content`,`images`,`status`,`created_at`) VALUES (?,?,?,?,?,?)")).
		WithArgs(2, "title", "content", "[]", "open", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	r := r006Engine(t, db)
	body := `{"title":"title","content":"content"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/questions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r006Token(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create question must not panic, status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestQuestionListEmptyArray(t *testing.T) {
	db, mock := r006DB(t)
	mock.ExpectQuery(r006CountRe).WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))
	mock.ExpectQuery(r006ListRe).WithArgs(10).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	r := r006Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/questions", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status=%d", rr.Code)
	}
	var body struct {
		Data struct {
			List json.RawMessage `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rr.Body.String())
	}
	if string(body.Data.List) == "null" {
		t.Fatalf("empty question list must be [] not null, body=%s", rr.Body.String())
	}
}
