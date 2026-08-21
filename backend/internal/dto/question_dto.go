package dto

// QuestionCreateRequest creates a Q&A question.
type QuestionCreateRequest struct {
	Title   string `json:"title" binding:"required,max=255"`
	Content string `json:"content" binding:"required"`
	Images  string `json:"images"`
}

// AnswerCreateRequest creates an reply.
type AnswerCreateRequest struct {
	Content string `json:"content" binding:"required"`
}

// AdoptRequest marks an reply as best.
type AdoptRequest struct {
	AnswerID uint `json:"answer_id" binding:"required"`
}
