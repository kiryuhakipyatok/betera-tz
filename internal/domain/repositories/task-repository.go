package repositories

import (
	"betera-tz/internal/domain/models"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/storage"
	"context"
	"errors"
	"fmt"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) (*string, error)
	GetById(ctx context.Context, id string) (*models.Task, error)
	Get(ctx context.Context, amount, page int, statusFilter string) ([]models.Task, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

type taskRepository struct {
	Storage *storage.Storage
}

func NewTaskRepository(s *storage.Storage) TaskRepository {
	return &taskRepository{
		Storage: s,
	}
}

const place = "taskRepository."

func (tr *taskRepository) Create(ctx context.Context, task *models.Task) (*string, error) {
	op := place + "Create"
	query := "INSERT INTO tasks (id, title, description, status) VALUES ($1,$2,$3,$4)"
	res, err := tr.Storage.Pool.Exec(ctx, query, task.ID, task.Title, task.Description, task.Status)
	if err != nil {
		if storage.ErrorAlreadyExists(err) {
			return nil, errs.ErrAlreadyExists(op, err)
		}
		return nil, errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return nil, errs.ErrNotFound(op)
	}
	taskId := task.ID.String()
	return &taskId, nil
}

func (tr *taskRepository) GetById(ctx context.Context, id string) (*models.Task, error) {
	op := place + "GetById"
	query := "SELECT * FROM tasks WHERE id = $1"
	task := models.Task{}
	if err := tr.Storage.Pool.QueryRow(ctx, query, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return &task, nil
}

func (tr *taskRepository) Get(ctx context.Context, amount, page int, statusFilter string) ([]models.Task, error) {
	op := place + "Get"
	args := []any{}
	strs := []string{}
	validStatus := map[string]bool{
		"created":    true,
		"processing": true,
		"done":       true,
	}
	i := 0
	if validStatus[statusFilter] {
		i++
		strs = append(strs, fmt.Sprintf("status = $%d", i))
		args = append(args, statusFilter)
	}
	query := "SELECT * FROM tasks"
	if len(strs) > 0 {
		query += " WHERE " + strs[0]
	}
	if amount > 0 && page > 0 {
		i++
		offset := (page - 1) * amount
		query += fmt.Sprintf(" OFFSET $%d LIMIT $%d", i, i+1)
		args = append(args, offset, amount)
	}

	tasks := []models.Task{}
	fmt.Println(query)
	rows, err := tr.Storage.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (tr *taskRepository) UpdateStatus(ctx context.Context, id, status string) error {
	op := place + "UpdateStatus"
	query := "UPDATE tasks SET status = $1 WHERE id = $2"
	res, err := tr.Storage.Pool.Exec(ctx, query, status, id)
	if err != nil {
		if storage.CheckErr(err) {
			return errs.ErrInvalidValues(op, err)
		}
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
