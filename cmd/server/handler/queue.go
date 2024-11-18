package handler

import (
	"httpqueue/lib/queue"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type QueueHandler struct {
	queues             map[string]*queue.Queue
	mu                 sync.RWMutex
	longPollingTimeout int
}

func NewQueueHandler(longPollingTimeout int) *QueueHandler {
	return &QueueHandler{
		queues:             make(map[string]*queue.Queue),
		longPollingTimeout: longPollingTimeout,
	}
}

func (qh *QueueHandler) getOrCreateQueue(name string) *queue.Queue {
	qh.mu.Lock()
	defer qh.mu.Unlock()

	q, exists := qh.queues[name]
	if !exists {
		q = queue.New()
		qh.queues[name] = q
	}
	return q
}

func (qh *QueueHandler) HandlePush(w http.ResponseWriter, r *http.Request) {

	queueName := r.PathValue("queueName")

	log.Printf("PUSH on %s\n", queueName)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}

	message := string(body)
	if strings.Contains(message, "\n") {
		http.Error(w, "Message cannot contain newlines", http.StatusBadRequest)
		return
	}

	queue := qh.getOrCreateQueue(queueName)
	queue.Push(message)
	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) HandlePull(w http.ResponseWriter, r *http.Request) {
	queueName := r.PathValue("queueName")

	if queueName == "" {
		http.Error(w, "Queue name required", http.StatusBadRequest)
		return
	}

	log.Printf("PULL on %s\n", queueName)

	queue := qh.getOrCreateQueue(queueName)

	//Implement long polling of 10 seconds
	deadline := time.Now().Add(time.Duration(qh.longPollingTimeout) * time.Second)
	for time.Now().Before(deadline) {
		if err, message := queue.Pull(); err == nil {
			w.Write([]byte(message))
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	w.WriteHeader(http.StatusNoContent)
}
