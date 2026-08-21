package service_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync"
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

func r002DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), SkipDefaultTransaction: false})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	return db, mock
}

func r002Engine(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	cfg := config.Load()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return router.Setup(cfg, db, l)
}

func r002Token(t *testing.T) string {
	t.Helper()
	cfg := config.Load()
	tok, err := util.GenerateToken(2, "gardener", "user", cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

var r002SelRe = regexp.QuoteMeta("SELECT * FROM `answers` WHERE `answers`.`id` = ? ORDER BY `answers`.`id` LIMIT ?")
var r002ListRe = regexp.QuoteMeta("SELECT * FROM `answers` WHERE question_id = ? ORDER BY is_best DESC, like_count DESC, id ASC")

func r002AnswerRows(likeCount int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "question_id", "user_id", "content", "is_best", "like_count", "created_at"}).
		AddRow(1, 1, 2, "content-1", false, likeCount, time.Now())
}

func TestAnswerLikeConcurrentRace(t *testing.T) {
	db, mock := r002DB(t)
	const n = 20
	for i := 0; i < n; i++ {
		mock.ExpectQuery(r002SelRe).WithArgs(1, 1).WillReturnRows(r002AnswerRows(5))
	}
	r := r002Engine(t, db)
	token := r002Token(t)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPut, "/api/v1/answers/1/like", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("like status=%d", rr.Code)
			}
		}()
	}
	close(start)
	wg.Wait()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAnswerLikeResponseCount(t *testing.T) {
	db, mock := r002DB(t)
	mock.ExpectQuery(r002SelRe).WithArgs(1, 1).WillReturnRows(r002AnswerRows(5))
	mock.ExpectQuery(r002SelRe).WithArgs(1, 1).WillReturnRows(r002AnswerRows(5))
	r := r002Engine(t, db)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/answers/1/like", nil)
	req.Header.Set("Authorization", "Bearer "+r002Token(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("like status=%d body=%s", rr.Code, rr.Body.String())
	}
	var body struct {
		Data struct {
			LikeCount int `json:"like_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data.LikeCount != 6 {
		t.Fatalf("like_count=%d want 6 (base 5 + 1 hot delta), body=%s", body.Data.LikeCount, rr.Body.String())
	}
}

func TestAnswerListMergesHotLikes(t *testing.T) {
	db, mock := r002DB(t)
	mock.ExpectQuery(r002SelRe).WithArgs(1, 1).WillReturnRows(r002AnswerRows(5))
	mock.ExpectQuery(r002ListRe).WithArgs(1).WillReturnRows(r002AnswerRows(5))
	r := r002Engine(t, db)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/answers/1/like", nil)
	req.Header.Set("Authorization", "Bearer "+r002Token(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("like status=%d", rr.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/questions/1/answers", nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("list status=%d", rr2.Code)
	}
	var body struct {
		Data []struct {
			ID        uint `json:"id"`
			LikeCount int  `json:"like_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal list: %v body=%s", err, rr2.Body.String())
	}
	if len(body.Data) != 1 || body.Data[0].ID != 1 || body.Data[0].LikeCount != 6 {
		t.Fatalf("list like_count mismatch: %+v", body.Data)
	}
}
