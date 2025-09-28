package workers

import (
	"betera-tz/internal/domain/repositories"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/logger"
	"betera-tz/pkg/queue"
	"context"
	"fmt"
	"time"
)

type TaskWorker struct {
	Consumer       *queue.Consumer
	Producer       *queue.Producer
	Logger         *logger.Logger
	TaskRepository repositories.TaskRepository
}

func NewTaskWorker(c *queue.Consumer, p *queue.Producer, l *logger.Logger, tr repositories.TaskRepository) *TaskWorker {
	return &TaskWorker{
		Consumer:       c,
		Producer:       p,
		Logger:         l,
		TaskRepository: tr,
	}
}

func (tw *TaskWorker) MustStart() {
	op := "TaskWorker.Start"
	log := tw.Logger.AddOp(op)
	log.Info("starting task worker")

	handler := func(message queue.Message) error {
		taskId := string(message.Value)
		return tw.processTask(context.Background(), taskId)
	}

	if err := tw.Consumer.HandleMessages(handler); err != nil {
		panic(fmt.Errorf("failed to start task worker: %w", err))
	}
}

func (tw *TaskWorker) processTask(ctx context.Context, id string) error {
	op := "worker.TaskProcessing"
	log := tw.Logger.AddOp(op)
	log.Info("task processing")
	if err := tw.TaskRepository.UpdateStatus(ctx, id, "processing"); err != nil {
		log.Error("failed to update task's status", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	time.Sleep(time.Second * 10)
	if err := tw.TaskRepository.UpdateStatus(ctx, id, "done"); err != nil {
		log.Error("failed to update task's status", logger.Err(err))
		return errs.NewAppError(op, err)
	}
	log.Info("task processed: ", "id", id)
	return nil
}
