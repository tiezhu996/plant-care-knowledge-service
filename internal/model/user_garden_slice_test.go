package model_test

import (
	"log/slog"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
)

func r005DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

var r005ListRe = regexp.QuoteMeta("SELECT * FROM `user_gardens` WHERE user_id = ? ORDER BY id DESC")

func TestGardenListContentExact(t *testing.T) {
	db, mock := r005DB(t)
	svc := service.NewUserGardenService(repository.NewUserGardenRepository(db), slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	rows := sqlmock.NewRows([]string{"id", "user_id", "plant_species_id", "nickname", "owned_since", "location", "care_reminder_id", "created_at"}).
		AddRow(1, 2, 10, "a", time.Now(), "", 0, time.Now()).
		AddRow(2, 2, 20, "b", time.Now(), "", 99, time.Now()).
		AddRow(3, 2, 30, "c", time.Now(), "", 0, time.Now()).
		AddRow(4, 2, 40, "d", time.Now(), "", 88, time.Now())
	mock.ExpectQuery(r005ListRe).WithArgs(2).WillReturnRows(rows)
	items, err := svc.List(2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("garden list must keep all 4 plants, got %d: %+v", len(items), items)
	}
	for i, want := range []uint{1, 2, 3, 4} {
		if items[i].ID != want {
			t.Fatalf("items[%d].ID=%d want %d (full list corrupted: %+v)", i, items[i].ID, want, items)
		}
	}
}
