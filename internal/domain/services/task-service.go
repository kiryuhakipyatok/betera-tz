package services

import (
	"betera-tz/internal/domain/models"
	"betera-tz/internal/domain/repositories"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/logger"
	"betera-tz/pkg/queue"
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, title, description string) (*uuid.UUID, error)
	GetById(ctx context.Context, id string) (*models.Task, error)
	Get(ctx context.Context, amount, page int, statusFilter string) ([]models.Task, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

type MessageProducer interface {
	SendMessage(message queue.Message) error
}

type taskService struct {
	TaskRepository repositories.TaskRepository
	Producer       MessageProducer
	Logger         *logger.Logger
}

func NewTaskService(tr repositories.TaskRepository, l *logger.Logger, p *queue.Producer) TaskService {
	return &taskService{
		TaskRepository: tr,
		Producer:       p,
		Logger:         l,
	}
}

const place = "taskService."

func (ts *taskService) Create(ctx context.Context, title, description string) (*uuid.UUID, error) {
	op := place + "Create"
	log := ts.Logger.AddOp(op)
	log.Info("creating task")
	task := &models.Task{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      "created",
	}
	id, err := ts.TaskRepository.Create(ctx, task)
	if err != nil {
		log.Error("failed to create task", logger.Err(err))
		return nil, errs.NewAppError(op, err)
	}

	msg := queue.Message{
		Key:   *id,
		Value: []byte(*id),
		Time:  time.Now(),
	}
	if err := ts.Producer.SendMessage(msg); err != nil {
		log.Error("failed to send task to queue", logger.Err(err))
	} else {
		log.Info("task sent to queue", "task_id", *id)
	}

	uid, err := uuid.Parse(*id)
	if err != nil {
		log.Error("failed to parse task id", logger.Err(err))
	}

	log.Info("task created")
	return &uid, nil
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

func (ts *taskService) Get(ctx context.Context, amount, page int, statusFilter string) ([]models.Task, error) {
	op := place + "Get"
	log := ts.Logger.AddOp(op)
	log.Info("fetching tasks")
	if amount <= 0 || page <= 0 {
		page = -1
		amount = -1
	}
	tasks, err := ts.TaskRepository.Get(ctx, amount, page, statusFilter)
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
		log.Error("failed to update task's status", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Info("task's status updated")
	return nil
}
