package service

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
)

func r002SvcDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func r002Svc(t *testing.T, db *gorm.DB) *AnswerService {
	t.Helper()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewAnswerService(db, repository.NewAnswerRepository(db), repository.NewQuestionRepository(db), l)
}

var r002SvcSelRe = regexp.QuoteMeta("SELECT * FROM `answers` WHERE `answers`.`id` = ? ORDER BY `answers`.`id` LIMIT ?")
var r002DeltaRe = regexp.QuoteMeta("UPDATE `answers` SET `like_count`=like_count + ? WHERE id = ?")

func TestAnswerFlushPersistAndIdempotent(t *testing.T) {
	db, mock := r002SvcDB(t)
	svc := r002Svc(t, db)
	for i := 0; i < 3; i++ {
		rows := sqlmock.NewRows([]string{"id", "question_id", "user_id", "content", "is_best", "like_count", "created_at"}).
			AddRow(1, 1, 2, "c", false, 5, time.Now())
		mock.ExpectQuery(r002SvcSelRe).WithArgs(1, 1).WillReturnRows(rows)
	}
	mock.ExpectBegin()
	mock.ExpectExec(r002DeltaRe).WithArgs(3, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	for i := 0; i < 3; i++ {
		if _, err := svc.Like(1); err != nil {
			t.Fatalf("like %d: %v", i, err)
		}
	}
	if err := svc.FlushLikes(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if err := svc.FlushLikes(); err != nil {
		t.Fatalf("second flush must be a no-op: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
