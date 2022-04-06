package operator

import (
	"context"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

const maxRetries = 5
const handleTimeout = time.Second
const workerFrequency = time.Second

type Runner interface {
	Process(e interface{})
	Run(ctx context.Context, h func(context.Context, interface{}) error)
}

type runner struct {
	queue           workqueue.RateLimitingInterface
	handle          func(context.Context, interface{}) error
	mutex           sync.RWMutex
	workerFrequency time.Duration
}

// NewRunner instantiates queue producer and consumer
func NewRunner() Runner {
	return &runner{
		queue:           workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		workerFrequency: workerFrequency,
	}
}

// Process adds entry to the processing queue
func (c *runner) Process(e interface{}) {
	c.queue.Add(e)
}

// Run will start ticker worker that will call handler func on each match
func (c *runner) Run(ctx context.Context, h func(context.Context, interface{}) error) {
	defer c.queue.ShutDown()

	c.mutex.Lock()
	c.handle = h
	c.mutex.Unlock()

	wait.UntilWithContext(ctx, c.worker, c.workerFrequency)
}

func (c *runner) worker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *runner) processNextItem(ctx context.Context) bool {
	e, quit := c.queue.Get()
	if quit {
		log.Error("Queue goes down!")
		return false
	}
	defer c.queue.Done(e)

	ctx, cancel := context.WithTimeout(ctx, handleTimeout)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.handle == nil {
		log.Fatal("no handler defined")
		return false
	}

	err := c.handle(ctx, e)
	if err == nil {
		c.queue.Forget(e)
		return true
	}

	if c.queue.NumRequeues(e) < maxRetries {
		log.Errorf("Error processing ev %v, retry. Error: %v", e, err)
		c.queue.AddRateLimited(e)
		return true
	}

	log.Errorf("Error processing %v Max retries achieved: %v", e, err)
	c.queue.Forget(e)
	utilruntime.HandleError(err)

	return true
}
