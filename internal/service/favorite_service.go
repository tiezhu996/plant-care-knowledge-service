package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// FavoriteService implements favorite/unfavorite logic.
type FavoriteService struct {
	repo   *repository.FavoriteRepository
	logger *slog.Logger
}

// NewFavoriteService creates a FavoriteService.
func NewFavoriteService(repo *repository.FavoriteRepository, logger *slog.Logger) *FavoriteService {
	return &FavoriteService{repo: repo, logger: logger}
}

// Add favorites a target for a user.
func (s *FavoriteService) Add(userID uint, targetType string, targetID uint) (*model.Favorite, error) {
	if !constants.IsValidFavoriteTarget(targetType) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("Favorite[target_type=%s] add failed: invalid target type", targetType))
	}
	f := &model.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID}
	if err := s.repo.Create(f); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, util.NewAppError(409, constants.CodeConflict,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] add failed: already favorited", userID, targetType, targetID))
		}
		s.logger.Error("favorite add failed", "error", err)
		return nil, fmt.Errorf("favorite add: %w", err)
	}
	s.logger.Info("favorite added", "user_id", userID, "target_type", targetType, "target_id", targetID)
	return f, nil
}

// Remove deletes a favorite.
func (s *FavoriteService) Remove(userID uint, targetType string, targetID uint) error {
	f, err := s.repo.Find(userID, targetType, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("Favorite[user_id=%d target=%s:%d] not found", userID, targetType, targetID))
		}
		return fmt.Errorf("favorite remove find: %w", err)
	}
	if err := s.repo.Delete(f.ID); err != nil {
		return fmt.Errorf("favorite remove: %w", err)
	}
	s.logger.Info("favorite removed", "user_id", userID, "target_type", targetType, "target_id", targetID)
	return nil
}

// ListByUser returns a user's favorites, retrying transient store errors.
func (s *FavoriteService) ListByUser(ctx context.Context, userID uint, targetType string) ([]model.Favorite, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		items, err := s.repo.ListByUser(ctx, userID, targetType)
		if err == nil {
			return items, nil
		}
		// A canceled context is not a transient store error: stop retrying so
		// client aborts/timeouts surface promptly instead of masking the
		// failure behind fresh background contexts.
		if ctx.Err() != nil {
			return nil, err
		}
		lastErr = err
	}
	return nil, fmt.Errorf("favorite list: %w", lastErr)
}
