package service

import (
	"context"
	"errors"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrCommentEmpty           = errors.New("comment cannot be empty")
	ErrCommentNotFound        = errors.New("comment not found")
	ErrCommentPermissioDenied = errors.New("permission denied to edit/delete comment")
)

type ExpenseCommentService interface {
	CreateComment(ctx context.Context, expenseID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error)
	ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error)
	UpdateComment(ctx context.Context, commentID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error)
	DeleteComment(ctx context.Context, commentID, userID pgtype.UUID) error
}

type expenseCommentService struct {
	repo            repository.ExpenseCommentRepository
	expenseService  ExpenseService
	activityService GroupActivityService
}

func NewExpenseCommentService(
	repo repository.ExpenseCommentRepository,
	expenseService ExpenseService,
	activityService GroupActivityService,
) ExpenseCommentService {
	return &expenseCommentService{
		repo:            repo,
		expenseService:  expenseService,
		activityService: activityService,
	}
}

func (s *expenseCommentService) CreateComment(ctx context.Context, expenseID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
	if comment == "" {
		return sqlc.ExpenseComment{}, ErrCommentEmpty
	}

	// Verify expense exists and get group ID (using GetExpenseByID which likely checks permissions too, or just existence)
	expense, err := s.expenseService.GetExpenseByID(ctx, expenseID, userID)
	if err != nil {
		return sqlc.ExpenseComment{}, err
	}

	result, err := s.repo.CreateComment(ctx, sqlc.CreateExpenseCommentParams{
		ExpenseID: expenseID,
		UserID:    userID,
		Comment:   comment,
	})
	if err != nil {
		return sqlc.ExpenseComment{}, err
	}

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    expense.GroupID,
		UserID:     userID,
		Action:     "comment_added",
		EntityType: "expense",
		EntityID:   expenseID,
		Metadata: map[string]interface{}{
			"comment_id":      result.ID,
			"comment_snippet": truncateString(comment, 50),
		},
	})

	return result, nil
}

func (s *expenseCommentService) ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
	return s.repo.ListComments(ctx, expenseID)
}

func (s *expenseCommentService) UpdateComment(ctx context.Context, commentID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
	if comment == "" {
		return sqlc.ExpenseComment{}, ErrCommentEmpty
	}

	// Check existance and ownership
	existing, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return sqlc.ExpenseComment{}, err
	}

	if existing.UserID != userID {
		return sqlc.ExpenseComment{}, ErrCommentPermissioDenied
	}

	return s.repo.UpdateComment(ctx, sqlc.UpdateExpenseCommentParams{
		ID:      commentID,
		Comment: comment,
	})
}

func (s *expenseCommentService) DeleteComment(ctx context.Context, commentID, userID pgtype.UUID) error {
	// Check existance and ownership
	existing, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return ErrCommentPermissioDenied
	}

	return s.repo.DeleteComment(ctx, commentID)
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
