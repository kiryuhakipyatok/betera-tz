package workers

import (
	"betera-tz/pkg/logger"
	"betera-tz/pkg/queue"
)

type Worker struct {
	Consumer *queue.Consumer
	Producer *queue.Producer
	Logger   *logger.Logger
}

func NewWorker(c *queue.Consumer, p *queue.Producer, l *logger.Logger) *Worker {
	return &Worker{
		Consumer: c,
		Producer: p,
		Logger:   l,
	}
}

func (w *Worker) Start(handler queue.MessageHandler) error {
	op := "Worker.Start"
	log := w.Logger.AddOp(op)
	log.Info("starting worker")

	return w.Consumer.HandleMessages(handler)
}
