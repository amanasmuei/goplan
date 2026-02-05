package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/domain/user"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// CommentRepository implements repository.CommentRepository using PostgreSQL.
type CommentRepository struct {
	pool *pgxpool.Pool
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// Create creates a new comment.
func (r *CommentRepository) Create(ctx context.Context, c *task.Comment) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	mentions := c.Mentions
	if mentions == nil {
		mentions = []string{}
	}

	result, err := q.CreateComment(ctx, sqlc.CreateCommentParams{
		TaskID:   c.TaskID,
		UserID:   c.UserID,
		Content:  c.Content,
		Mentions: mentions,
	})
	if err != nil {
		return MapError(err, "comment")
	}

	c.ID = result.ID
	c.CreatedAt = result.CreatedAt.Time
	c.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a comment by ID.
func (r *CommentRepository) GetByID(ctx context.Context, id string) (*task.Comment, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetCommentByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "comment")
	}

	return sqlcCommentRowToDomain(result), nil
}

// Update updates comment content.
func (r *CommentRepository) Update(ctx context.Context, id string, content string, mentions []string) (*task.Comment, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	if mentions == nil {
		mentions = []string{}
	}

	result, err := q.UpdateComment(ctx, sqlc.UpdateCommentParams{
		ID:       id,
		Content:  content,
		Mentions: mentions,
	})
	if err != nil {
		return nil, MapError(err, "comment")
	}

	// Get full comment with user info
	fullComment, err := q.GetCommentByID(ctx, id)
	if err != nil {
		// Fall back to basic result
		return &task.Comment{
			ID:        result.ID,
			TaskID:    result.TaskID,
			UserID:    result.UserID,
			Content:   result.Content,
			Mentions:  result.Mentions,
			CreatedAt: result.CreatedAt.Time,
			UpdatedAt: result.UpdatedAt.Time,
		}, nil
	}

	return sqlcCommentRowToDomain(fullComment), nil
}

// Delete deletes a comment by ID.
func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeleteComment(ctx, id)
	if err != nil {
		return MapError(err, "comment")
	}

	return nil
}

// ListByTask retrieves all comments for a task with user info.
func (r *CommentRepository) ListByTask(ctx context.Context, taskID string) ([]*task.Comment, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	comments, err := q.ListCommentsByTask(ctx, taskID)
	if err != nil {
		return nil, MapError(err, "comment")
	}

	result := make([]*task.Comment, len(comments))
	for i, c := range comments {
		result[i] = sqlcCommentListRowToDomain(c)
	}

	return result, nil
}

// ListByTaskPaginated retrieves paginated comments for a task.
func (r *CommentRepository) ListByTaskPaginated(ctx context.Context, taskID string, pagination repository.Pagination) (*repository.PaginatedResult[task.Comment], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	comments, err := q.ListCommentsByTaskPaginated(ctx, sqlc.ListCommentsByTaskPaginatedParams{
		TaskID: taskID,
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "comment")
	}

	count, err := q.CountCommentsByTask(ctx, taskID)
	if err != nil {
		return nil, MapError(err, "comment")
	}

	items := make([]task.Comment, len(comments))
	for i, c := range comments {
		items[i] = *sqlcCommentPaginatedRowToDomain(c)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// CountByTask counts comments for a task.
func (r *CommentRepository) CountByTask(ctx context.Context, taskID string) (int, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	count, err := q.CountCommentsByTask(ctx, taskID)
	if err != nil {
		return 0, MapError(err, "comment")
	}

	return int(count), nil
}

// Helper functions

func sqlcCommentRowToDomain(c sqlc.GetCommentByIDRow) *task.Comment {
	mentions := c.Mentions
	if mentions == nil {
		mentions = []string{}
	}

	comment := &task.Comment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		Mentions:  mentions,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}

	// Add user info if available
	if c.UserName != "" {
		comment.User = &user.User{
			ID:        c.UserID,
			Name:      c.UserName,
			Email:     c.UserEmail,
			AvatarURL: textToPtr(c.UserAvatar),
		}
	}

	return comment
}

func sqlcCommentListRowToDomain(c sqlc.ListCommentsByTaskRow) *task.Comment {
	mentions := c.Mentions
	if mentions == nil {
		mentions = []string{}
	}

	comment := &task.Comment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		Mentions:  mentions,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}

	// Add user info
	if c.UserName != "" {
		comment.User = &user.User{
			ID:        c.UserID,
			Name:      c.UserName,
			Email:     c.UserEmail,
			AvatarURL: textToPtr(c.UserAvatar),
		}
	}

	return comment
}

func sqlcCommentPaginatedRowToDomain(c sqlc.ListCommentsByTaskPaginatedRow) *task.Comment {
	mentions := c.Mentions
	if mentions == nil {
		mentions = []string{}
	}

	comment := &task.Comment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		Mentions:  mentions,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}

	// Add user info
	if c.UserName != "" {
		comment.User = &user.User{
			ID:        c.UserID,
			Name:      c.UserName,
			Email:     c.UserEmail,
			AvatarURL: textToPtr(c.UserAvatar),
		}
	}

	return comment
}
