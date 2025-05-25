package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type TaskStatus string

type Task struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	Position  float64    `json:"position"`
	ProjectID uuid.UUID  `json:"project_id"`
}

// TaskStore interface for position-related operations
type TaskStore interface {
	CountItems(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (int, error)
	GetFirstPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error)
	GetLastPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error)
	GetPositions(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID, offset, limit int) ([]float64, error)
}

// PostgresTaskStore implementation
type PostgresTaskStore struct {
	tx pgx.Tx
}

func NewPostgresTaskStore(tx pgx.Tx) *PostgresTaskStore {
	return &PostgresTaskStore{tx: tx}
}

func (s *PostgresTaskStore) CountItems(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM issues 
		WHERE project_id = $1 AND status = $2 AND id != $3
	`
	err := s.tx.QueryRow(ctx, query, projectID, status, excludeID).Scan(&count)
	return count, err
}

func (s *PostgresTaskStore) GetFirstPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error) {
	var position float64
	query := `
		SELECT position 
		FROM issues 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY position ASC 
		LIMIT 1
	`
	err := s.tx.QueryRow(ctx, query, projectID, status, excludeID).Scan(&position)
	return position, err
}

func (s *PostgresTaskStore) GetLastPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error) {
	var position float64
	query := `
		SELECT position 
		FROM issues 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY position DESC 
		LIMIT 1
	`
	err := s.tx.QueryRow(ctx, query, projectID, status, excludeID).Scan(&position)
	return position, err
}

func (s *PostgresTaskStore) GetPositions(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID, offset, limit int) ([]float64, error) {
	query := `
		SELECT position 
		FROM issues 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY position ASC 
		LIMIT $4 OFFSET $5
	`

	rows, err := s.tx.Query(ctx, query, projectID, status, excludeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []float64
	for rows.Next() {
		var pos float64
		if err := rows.Scan(&pos); err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	return positions, rows.Err()
}

// Position calculation function - now pure and testable
func CalculateNewPosition(ctx context.Context, store TaskStore, projectID uuid.UUID, status TaskStatus, targetIndex int, excludeID uuid.UUID) (float64, error) {
	count, err := store.CountItems(ctx, projectID, status, excludeID)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}

	if count == 0 {
		return 1000.0, nil
	}

	if targetIndex <= 0 {
		// Insert at beginning
		firstPos, err := store.GetFirstPosition(ctx, projectID, status, excludeID)
		if err != nil {
			return 0, fmt.Errorf("failed to get first position: %w", err)
		}
		return firstPos - 1000.0, nil
	}

	if targetIndex >= count {
		// Insert at end
		lastPos, err := store.GetLastPosition(ctx, projectID, status, excludeID)
		if err != nil {
			return 0, fmt.Errorf("failed to get last position: %w", err)
		}
		return lastPos + 1000.0, nil
	}

	// Insert between two positions
	positions, err := store.GetPositions(ctx, projectID, status, excludeID, targetIndex-1, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to get positions: %w", err)
	}

	if len(positions) < 2 {
		return 0, fmt.Errorf("insufficient positions returned")
	}

	return (positions[0] + positions[1]) / 2.0, nil
}

// Move request struct
type MoveTaskRequest struct {
	TaskID      uuid.UUID  `json:"issue_id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	NewStatus   TaskStatus `json:"new_status"`
	NewPosition int        `json:"new_position"`
}

// Handler implementation
type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) MoveTask(ctx context.Context, req MoveTaskRequest) (*Task, error) {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	store := NewPostgresTaskStore(tx)

	// Calculate new position
	newPosition, err := CalculateNewPosition(ctx, store, req.ProjectID, req.NewStatus, req.NewPosition, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate new position: %w", err)
	}

	// Update issue
	var updatedTask Task
	updateQuery := `
		UPDATE issues 
		SET status = $1, position = $2, updated_at = NOW() 
		WHERE id = $3 AND project_id = $4
		RETURNING id, title, status, position, project_id
	`

	err = tx.QueryRow(ctx, updateQuery, req.NewStatus, newPosition, req.TaskID, req.ProjectID).Scan(
		&updatedTask.ID,
		&updatedTask.Title,
		&updatedTask.Status,
		&updatedTask.Position,
		&updatedTask.ProjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedTask, nil
}

// Mock implementation for testing
type MockTaskStore struct {
	CountItemsFunc       func(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (int, error)
	GetFirstPositionFunc func(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error)
	GetLastPositionFunc  func(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error)
	GetPositionsFunc     func(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID, offset, limit int) ([]float64, error)
}

func (m *MockTaskStore) CountItems(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (int, error) {
	if m.CountItemsFunc != nil {
		return m.CountItemsFunc(ctx, projectID, status, excludeID)
	}
	return 0, nil
}

func (m *MockTaskStore) GetFirstPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error) {
	if m.GetFirstPositionFunc != nil {
		return m.GetFirstPositionFunc(ctx, projectID, status, excludeID)
	}
	return 0, nil
}

func (m *MockTaskStore) GetLastPosition(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID) (float64, error) {
	if m.GetLastPositionFunc != nil {
		return m.GetLastPositionFunc(ctx, projectID, status, excludeID)
	}
	return 0, nil
}

func (m *MockTaskStore) GetPositions(ctx context.Context, projectID uuid.UUID, status TaskStatus, excludeID uuid.UUID, offset, limit int) ([]float64, error) {
	if m.GetPositionsFunc != nil {
		return m.GetPositionsFunc(ctx, projectID, status, excludeID, offset, limit)
	}
	return nil, nil
}
