package queue

import (
	"errors"
	"sync"
)

type Queue struct {
	messages []string
	mu       sync.RWMutex
}

func New() *Queue {
	return &Queue{
		messages: make([]string, 0),
	}
}

func (q *Queue) Push(message string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, message)
}

func (q *Queue) Pull() (error, string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// If there are messages in the buffer, we send the last one right away
	if len(q.messages) > 0 {
		message := q.messages[0]
		q.messages = q.messages[1:]
		return nil, message
	}

	return errors.New("No elements in queue"), ""
}
