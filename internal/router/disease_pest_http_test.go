package router_test

import (
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

func r009DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func r009Engine(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	cfg := config.Load()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return router.Setup(cfg, db, l)
}

func r009AdminToken(t *testing.T) string {
	t.Helper()
	cfg := config.Load()
	tok, err := util.GenerateToken(1, "admin", "admin", cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

var r009SelRe = regexp.QuoteMeta("SELECT * FROM `disease_pests` WHERE `disease_pests`.`id` = ? ORDER BY `disease_pests`.`id` LIMIT ?")

func TestPestDetailGone(t *testing.T) {
	db, mock := r009DB(t)
	mock.ExpectQuery(r009SelRe).WithArgs(9999, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	r := r009Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pests/9999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET missing pest: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
}

func TestPestDeleteGone(t *testing.T) {
	db, mock := r009DB(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `disease_pests` WHERE `disease_pests`.`id` = ?")).
		WithArgs(9999).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	r := r009Engine(t, db)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/pests/9999", nil)
	req.Header.Set("Authorization", "Bearer "+r009AdminToken(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("DELETE missing pest: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
}

func TestPestEditGone(t *testing.T) {
	db, mock := r009DB(t)
	mock.ExpectQuery(r009SelRe).WithArgs(9999, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	r := r009Engine(t, db)
	body := `{"name":"新病虫害","symptoms":"黄叶"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/pests/9999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r009AdminToken(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("PUT missing pest: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
}
