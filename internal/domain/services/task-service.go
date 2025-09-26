package services

import (
	"betera-tz/internal/domain/models"
	"betera-tz/internal/domain/repositories"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/logger"
	"context"

	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, title, description string) error
	GetById(ctx context.Context, id string) (*models.Task, error)
	Get(ctx context.Context, amount, page int) ([]models.Task, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

type taskService struct {
	TaskRepository repositories.TaskRepository
	Logger         *logger.Logger
}

func NewTaskService(tr repositories.TaskRepository, l *logger.Logger) TaskService {
	return &taskService{
		TaskRepository: tr,
		Logger:         l,
	}
}

const place = "taskService."

func (ts *taskService) Create(ctx context.Context, title, description string) error {
	op := place + "Create"
	log := ts.Logger.AddOp(op)
	log.Info("creating task")
	task := &models.Task{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      "created",
	}
	if err := ts.TaskRepository.Create(ctx, task); err != nil {
		log.Error("failed to create task", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Info("task created")
	return nil
}

func (ts *taskService) GetById(ctx context.Context, id string) (*models.Task, error) {
	op := place + "GetById"
	log := ts.Logger.AddOp(op)
	log.Info("receiving task by id")
	task, err := ts.TaskRepository.GetById(ctx, id)
	if err != nil {
		log.Error("failed to receive task by id", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Info("task received")
	return task, nil
}

func (ts *taskService) Get(ctx context.Context, amount, page int) ([]models.Task, error) {
	op := place + "Get"
	log := ts.Logger.AddOp(op)
	log.Info("fetching tasks")
	if amount <= 0 || page <= 0 {
		page = 0
		amount = -1
	}
	tasks, err := ts.TaskRepository.Get(ctx, amount, page)
	if err != nil {
		log.Error("failed to fetch tasks", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}
	log.Info("tasks fetched")
	return tasks, nil
}

func (ts *taskService) UpdateStatus(ctx context.Context, id, status string) error {
	op := place + "UpdateStatus"
	log := ts.Logger.AddOp(op)
	log.Info("updating task's status")
	if err := ts.TaskRepository.UpdateStatus(ctx, id, status); err != nil {
		log.Error("failed to update task's status")
		return errs.NewAppError(op, err)
	}
	log.Info("task's status updated")
	return nil
}
