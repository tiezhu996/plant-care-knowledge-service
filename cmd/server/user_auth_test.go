package main_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/router"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func r010DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func r010Engine(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	cfg := config.Load()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return router.Setup(cfg, db, l)
}

func r010Token(t *testing.T, userID uint, role string) string {
	t.Helper()
	cfg := config.Load()
	tok, err := util.GenerateToken(userID, "u", role, cfg.JWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tok
}

func r010EvilIssuerToken(t *testing.T) string {
	t.Helper()
	cfg := config.Load()
	claims := util.Claims{
		UserID:   2,
		Username: "gardener",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "evil",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

var r010SelByIDRe = regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT ?")
var r010UpdRe = regexp.QuoteMeta("UPDATE `users` SET `username`=?,`email`=?,`password_hash`=?,`nickname`=?,`avatar`=?,`bio`=?,`role`=?,`created_at`=? WHERE `id` = ?")

func r010UserRow(id uint, role string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "nickname", "avatar", "bio", "role", "created_at"}).
		AddRow(id, "gardener", "gardener@gbplantwiki.local", "hash", "绿手指", "", "", role, time.Now())
}

func TestUserCannotSelfPromote(t *testing.T) {
	db, mock := r010DB(t)
	mock.ExpectQuery(r010SelByIDRe).WithArgs(2, 1).WillReturnRows(r010UserRow(2, "user"))
	mock.ExpectBegin()
	mock.ExpectExec(r010UpdRe).WithArgs("gardener", "gardener@gbplantwiki.local", "hash", "绿手指", "", "", "admin", sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	r := r010Engine(t, db)
	body := `{"nickname":"绿手指","role":"admin"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r010Token(t, 2, "user"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatalf("regular user must not self-promote to admin, got 200 body=%s", rr.Body.String())
	}
}

func TestStaleRoleRevoked(t *testing.T) {
	db, mock := r010DB(t)
	mock.ExpectQuery(r010SelByIDRe).WithArgs(2, 1).WillReturnRows(r010UserRow(2, "user"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `disease_pests` WHERE `disease_pests`.`id` = ?")).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	r := r010Engine(t, db)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/pests/1", nil)
	req.Header.Set("Authorization", "Bearer "+r010Token(t, 2, "admin"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("stale admin role in token must be revoked, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestTokenIssuerRejected(t *testing.T) {
	db, mock := r010DB(t)
	mock.ExpectQuery(r010SelByIDRe).WithArgs(2, 1).WillReturnRows(r010UserRow(2, "user"))
	r := r010Engine(t, db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+r010EvilIssuerToken(t))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatalf("token with wrong issuer must be rejected, got 200 body=%s", rr.Body.String())
	}
}

func TestLoginByEmail(t *testing.T) {
	db, mock := r010DB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.MinCost)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("gardener@gbplantwiki.local", 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "nickname", "avatar", "bio", "role", "created_at"}).
		AddRow(2, "gardener", "gardener@gbplantwiki.local", string(hash), "绿手指", "", "", "user", time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("gardener@gbplantwiki.local", 1).WillReturnRows(rows)
	r := r010Engine(t, db)
	body := `{"username":"gardener@gbplantwiki.local","password":"user123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login by email must succeed, got %d body=%s", rr.Code, rr.Body.String())
	}
}
