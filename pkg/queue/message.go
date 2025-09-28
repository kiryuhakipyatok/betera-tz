package queue

import (
	"time"
)

type Message struct {
	Key   string    `json:"key"`
	Value []byte    `json:"value"`
	Time  time.Time `json:"time"`
}

type MessageHandler func(message Message) error
