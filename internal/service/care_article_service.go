package service

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// CareArticleService implements care article business logic.
type CareArticleService struct {
	repo   *repository.CareArticleRepository
	logger *slog.Logger
	viewCh chan uint
	errCh  chan error
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewCareArticleService creates a CareArticleService.
func NewCareArticleService(repo *repository.CareArticleRepository, logger *slog.Logger) *CareArticleService {
	return &CareArticleService{repo: repo, logger: logger}
}

// Create adds an article owned by the current user.
func (s *CareArticleService) Create(userID uint, a *model.CareArticle) (*model.CareArticle, error) {
	if !constants.IsValidTopicTag(a.TopicTag) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("CareArticle[topic_tag=%s] create failed: invalid topic tag", a.TopicTag))
	}
	a.UserID = userID
	if a.Status == "" {
		a.Status = constants.ArticleStatusPublished
	}
	if err := s.repo.Create(a); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogArticleCreateFailed, a.Title), "error", err)
		return nil, fmt.Errorf("care article create: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogArticleCreateSuccess, a.Title), "id", a.ID, "user_id", userID)
	return a, nil
}

// Get returns an article and records a buffered view for the flush worker.
func (s *CareArticleService) Get(id uint) (*model.CareArticle, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("CareArticle[id=%d] not found", id))
		}
		return nil, fmt.Errorf("care article get: %w", err)
	}
	s.RecordView(id)
	return a, nil
}

// RecordView queues a view event for the background flusher.
func (s *CareArticleService) RecordView(id uint) {
	if s.viewCh == nil {
		return
	}
	s.viewCh <- id
}

// StartViewFlusher launches the background worker that batches view deltas.
func (s *CareArticleService) StartViewFlusher() {
	s.viewCh = make(chan uint)
	s.errCh = make(chan error)
	s.stopCh = make(chan struct{})
	go s.flushLoop()
}

func (s *CareArticleService) flushLoop() {
	s.wg.Add(1)
	defer s.wg.Done()
	pending := make(map[uint]int)
	for {
		select {
		case id := <-s.viewCh:
			pending[id]++
		case <-s.stopCh:
			for id := range pending {
				if err := s.repo.ApplyViewDelta(id, 1); err != nil {
					s.errCh <- err
				}
			}
			return
		}
	}
}

// StopViewFlusher asks the worker to stop and returns once it has exited.
func (s *CareArticleService) StopViewFlusher() error {
	close(s.stopCh)
	s.wg.Wait()
	return nil
}

// Update edits an article, verifying ownership.
func (s *CareArticleService) Update(id, userID uint, a *model.CareArticle) (*model.CareArticle, error) {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("care article update find: %w", err)
	}
	if exist.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden,
			fmt.Sprintf("CareArticle[id=%d] update failed: user_id=%d is not owner", id, userID))
	}
	if a.Title != "" {
		exist.Title = a.Title
	}
	if a.Content != "" {
		exist.Content = a.Content
	}
	if a.Cover != "" {
		exist.Cover = a.Cover
	}
	if a.TopicTag != "" {
		if !constants.IsValidTopicTag(a.TopicTag) {
			return nil, util.NewAppError(422, constants.CodeValidationError, "invalid topic tag")
		}
		exist.TopicTag = a.TopicTag
	}
	if err := s.repo.Update(exist); err != nil {
		return nil, fmt.Errorf("care article update: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogArticleUpdateSuccess, id), "id", id)
	return exist, nil
}

// Delete removes an article, verifying ownership.
func (s *CareArticleService) Delete(id, userID uint) error {
	exist, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("care article delete find: %w", err)
	}
	if exist.UserID != userID {
		return util.NewAppError(403, constants.CodeForbidden, fmt.Sprintf("CareArticle[id=%d] delete failed: not owner", id))
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("care article delete: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogArticleDeleteSuccess, id), "id", id)
	return nil
}

// List filters articles by topic tag and keyword.
func (s *CareArticleService) List(topicTag, keyword string, page, pageSize int) ([]model.CareArticle, int64, error) {
	items, total, err := s.repo.List(topicTag, keyword, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("care article list: %w", err)
	}
	return items, total, nil
}

// ListLatest returns the newest articles for the home page.
func (s *CareArticleService) ListLatest(limit int) ([]model.CareArticle, error) {
	items, err := s.repo.ListLatest(limit)
	if err != nil {
		return nil, fmt.Errorf("care article latest: %w", err)
	}
	return items, nil
}
