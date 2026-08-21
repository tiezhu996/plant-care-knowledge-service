package dto_test

import (
	"log/slog"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
)

func TestQuestionServiceListEmpty(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	db, _ := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), SkipDefaultTransaction: false})
	svc := service.NewQuestionService(repository.NewQuestionRepository(db), repository.NewAnswerRepository(db),
		service.NewUserService(repository.NewUserRepository(db), slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})), nil),
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `questions`")).WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `questions` ORDER BY id DESC LIMIT ?")).WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	items, _, err := svc.List(1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if items == nil {
		t.Fatalf("empty question list must be non-nil")
	}
}
