package util_test

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

	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
)

func r004DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func r004Svc(t *testing.T, db *gorm.DB) *service.CareReminderService {
	t.Helper()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewCareReminderService(repository.NewCareReminderRepository(db), l)
}

var r004SelRe = regexp.QuoteMeta("SELECT * FROM `care_reminders` WHERE `care_reminders`.`id` = ? ORDER BY `care_reminders`.`id` LIMIT ?")
var r004UpdRe = regexp.QuoteMeta("UPDATE `care_reminders` SET `status`=?,`updated_at`=? WHERE `id` = ?")

func TestReminderSkipDoneRejected(t *testing.T) {
	db, mock := r004DB(t)
	svc := r004Svc(t, db)
	mock.ExpectQuery(r004SelRe).WithArgs(1, 1).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "plant_species_id", "task_title", "remind_date", "frequency", "status", "created_at"}).
			AddRow(1, 1, 1, "task", time.Now().Add(-time.Hour), "weekly", model.ReminderDone, time.Now()))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `care_reminders` SET .* WHERE `id` = \\?").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), model.ReminderSkipped, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if _, err := svc.UpdateStatus(1, 1, model.ReminderSkipped); err == nil {
		t.Fatalf("skipping an already done reminder must be rejected")
	}
}

func TestReminderMarkOverdueExcludesSkipped(t *testing.T) {
	db, mock := r004DB(t)
	repo := repository.NewCareReminderRepository(db)
	goldRe := regexp.QuoteMeta("UPDATE `care_reminders` SET `status`=? WHERE user_id = ? AND status = ? AND remind_date < ?")
	mock.ExpectBegin()
	mock.ExpectExec(goldRe).WithArgs(model.ReminderOverdue, 1, model.ReminderPending, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if _, err := repo.MarkOverdue(1); err != nil {
		t.Fatalf("MarkOverdue should only touch pending reminders: %v", err)
	}
}

func TestReminderListExcludesSkipped(t *testing.T) {
	db, mock := r004DB(t)
	repo := repository.NewCareReminderRepository(db)
	goldRe := regexp.QuoteMeta("SELECT * FROM `care_reminders` WHERE user_id = ? AND status != ? ORDER BY remind_date ASC")
	mock.ExpectQuery(goldRe).WithArgs(1, model.ReminderSkipped).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "status", "remind_date"}).
			AddRow(1, 1, model.ReminderPending, time.Now()))
	if _, err := repo.ListByUser(1, ""); err != nil {
		t.Fatalf("list: %v", err)
	}
}
