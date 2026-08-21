package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func r003DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

var r003ListRe = regexp.QuoteMeta("SELECT * FROM `favorites` WHERE user_id = ? ORDER BY id DESC")

func TestFavoriteRepoCtxNotStored(t *testing.T) {
	db, mock := r003DB(t)
	repo := NewFavoriteRepository(db)

	ctx1, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListByUser(ctx1, 1, ""); err == nil {
		t.Fatalf("first call with canceled ctx should fail")
	}

	mock.ExpectQuery(r003ListRe).WithArgs(1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "target_type", "target_id", "created_at"}).
			AddRow(1, 1, "plant", 1, time.Now()))
	items, err := repo.ListByUser(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("second call with healthy ctx must not be polluted by the first: %v", err)
	}
	if len(items) != 1 || items[0].TargetID != 1 {
		t.Fatalf("unexpected items: %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
