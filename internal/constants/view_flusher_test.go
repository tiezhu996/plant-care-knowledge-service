package constants_test

import (
	"log/slog"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
)

func r008DB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func r008Svc(t *testing.T, db *gorm.DB) *service.CareArticleService {
	t.Helper()
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewCareArticleService(repository.NewCareArticleRepository(db), l)
}

var r008DeltaRe = regexp.QuoteMeta("UPDATE `care_articles` SET `view_count`=view_count + ? WHERE id = ?")

func TestViewFlusherAllViewsFlushed(t *testing.T) {
	db, mock := r008DB(t)
	svc := r008Svc(t, db)
	svc.StartViewFlusher()
	const n = 10
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			svc.RecordView(1)
		}()
	}
	close(start)
	wg.Wait()
	mock.ExpectBegin()
	mock.ExpectExec(r008DeltaRe).WithArgs(10, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	done := make(chan error, 1)
	go func() { done <- svc.StopViewFlusher() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("StopViewFlusher hung before flushing all views")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("all 10 views must be flushed: %v", err)
	}
}

func TestViewFlusherErrorBranchNoHang(t *testing.T) {
	db, mock := r008DB(t)
	svc := r008Svc(t, db)
	svc.StartViewFlusher()
	svc.RecordView(1)
	mock.ExpectBegin()
	mock.ExpectExec(r008DeltaRe).WithArgs(1, 1).WillReturnError(errFlushFailure)
	mock.ExpectRollback()
	done := make(chan error, 1)
	go func() { done <- svc.StopViewFlusher() }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("StopViewFlusher must surface the flush error")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("StopViewFlusher hung on flush error")
	}
}

func TestViewFlusherRecordAfterStop(t *testing.T) {
	db, _ := r008DB(t)
	svc := r008Svc(t, db)
	svc.StartViewFlusher()
	if err := svc.StopViewFlusher(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	done := make(chan struct{})
	go func() {
		svc.RecordView(2)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("RecordView must not block after the flusher stopped")
	}
}
