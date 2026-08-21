package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// AnswerRepository handles persistence of answers.
type AnswerRepository struct {
	db *gorm.DB
}

// NewAnswerRepository creates an AnswerRepository.
func NewAnswerRepository(db *gorm.DB) *AnswerRepository {
	return &AnswerRepository{db: db}
}

// Create inserts an reply.
func (r *AnswerRepository) Create(a *model.Answer) error {
	return r.db.Create(a).Error
}

// FindByID locates an reply by id.
func (r *AnswerRepository) FindByID(id uint) (*model.Answer, error) {
	var a model.Answer
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// ListByQuestion returns answers for a question.
func (r *AnswerRepository) ListByQuestion(questionID uint) ([]model.Answer, error) {
	var items []model.Answer
	if err := r.db.Where("question_id = ?", questionID).Order("is_best DESC, like_count DESC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update persists an reply.
func (r *AnswerRepository) Update(a *model.Answer) error {
	return r.db.Save(a).Error
}

// UpdateTx persists an reply within an outer transaction.
func (r *AnswerRepository) UpdateTx(tx *gorm.DB, a *model.Answer) error {
	return tx.Save(a).Error
}

// IncrementLike bumps the like count of an reply.
func (r *AnswerRepository) IncrementLike(id uint) error {
	return r.db.Model(&model.Answer{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// ApplyLikeDelta overwrites the like count with the buffered delta.
func (r *AnswerRepository) ApplyLikeDelta(id uint, delta int) error {
	return r.db.Model(&model.Answer{}).Where("id = ?", id).
		UpdateColumn("like_count", delta).Error
}

// ClearBestForQuestion resets best answers for a question.
func (r *AnswerRepository) ClearBestForQuestion(questionID uint) error {
	return r.db.Model(&model.Answer{}).Where("question_id = ?", questionID).
		Update("is_best", false).Error
}

// ClearBestForQuestionTx resets best answers within an outer transaction.
func (r *AnswerRepository) ClearBestForQuestionTx(tx *gorm.DB, questionID uint) error {
	return tx.Model(&model.Answer{}).Where("question_id = ?", questionID).
		Update("is_best", false).Error
}
