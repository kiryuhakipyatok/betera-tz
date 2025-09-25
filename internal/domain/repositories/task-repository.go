package repositories

import (
	"betera-tz/internal/domain/models"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/storage"
	"context"
	"errors"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetById(ctx context.Context, id string) (*models.Task, error)
	Get(ctx context.Context, amount, page int) ([]models.Task, error)
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

const space = "taskRepository."

func (tr *taskRepository) Create(ctx context.Context, task *models.Task) error {
	op := space + "Create"
	query := "INSERT INTO tasks (id, title, description, status) VALUES ($1,$2,$3,$4)"
	res, err := tr.Storage.Pool.Exec(ctx, query, task.ID, task.Title, task.Descpription, task.Status)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}

func (tr *taskRepository) GetById(ctx context.Context, id string) (*models.Task, error) {
	op := space + "GetById"
	query := "SELECT * FROM tasks WHERE id = $1"
	task := models.Task{}
	if err := tr.Storage.Pool.QueryRow(ctx, query, id).Scan(&task.ID, &task.Title, &task.Descpription, &task.Status); err != nil {
		if errors.Is(err, storage.ErrNotFound()) {
			return nil, errs.ErrNotFound(op)
		}
		return nil, errs.NewAppError(op, err)
	}
	return &task, nil
}

func (tr *taskRepository) Get(ctx context.Context, amount, page int) ([]models.Task, error) {
	op := space + "Get"
	query := "SELECT * FROM tasks OFFSET $1 LIMIT $2"
	tasks := []models.Task{}
	rows, err := tr.Storage.Pool.Query(ctx, query, amount*page-amount, amount)
	if err != nil {
		return nil, errs.NewAppError(op, err)
	}
	defer rows.Close()
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Descpription, &task.Status); err != nil {
			return nil, errs.NewAppError(op, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (tr *taskRepository) UpdateStatus(ctx context.Context, id, status string) error {
	op := space + "UpdateStatus"
	query := "UPDATE tasks SET status = $1 WHERE id = $2"
	res, err := tr.Storage.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return errs.NewAppError(op, err)
	}
	if res.RowsAffected() == 0 {
		return errs.ErrNotFound(op)
	}
	return nil
}
