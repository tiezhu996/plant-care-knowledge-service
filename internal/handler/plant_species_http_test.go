package handler_test

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

func r001DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	return db, mock
}

func r001Engine(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return router.Setup(cfg, db, logger)
}

func r001AdminToken(t *testing.T) string {
	t.Helper()
	cfg := config.Load()
	tok, err := util.GenerateToken(1, "admin", "admin", cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

func r001Code(t *testing.T, rr *httptest.ResponseRecorder) int {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err == nil {
		if code, ok := body["code"].(float64); ok {
			return int(code)
		}
	}
	return 0
}

func TestPlantDetailCase(t *testing.T) {
	db, mock := r001DB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `plant_species` WHERE `plant_species`.`id` = ? ORDER BY `plant_species`.`id` LIMIT ?")).
		WithArgs(9999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	r := r001Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plants/9999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET missing plant: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPlantDeleteCase(t *testing.T) {
	db, mock := r001DB(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `plant_species` WHERE `plant_species`.`id` = ?")).
		WithArgs(9999).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	r := r001Engine(t, db)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plants/9999", nil)
	req.Header.Set("Authorization", "Bearer "+r001AdminToken(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("DELETE missing plant: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPlantEditCase(t *testing.T) {
	db, mock := r001DB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `plant_species` WHERE `plant_species`.`id` = ? ORDER BY `plant_species`.`id` LIMIT ?")).
		WithArgs(9999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	r := r001Engine(t, db)
	body := `{"family":"天南星科","genus":"龟背竹属","name":"新植物","type":"foliage"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/plants/9999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r001AdminToken(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("PUT missing plant: status=%d want 404, body=%s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
