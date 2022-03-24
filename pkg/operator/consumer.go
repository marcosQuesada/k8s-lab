package operator

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

const defaultConciliationFrequency = time.Second * 5
const maxRequeue = 5

// EventProcessor handles event updates
type EventProcessor interface {
	Run(stopCh chan struct{})
	Handle(ctx context.Context, ev Event) error
}

type consumer struct {
	processor             EventProcessor
	queue                 workqueue.RateLimitingInterface
	conciliationFrequency time.Duration
}

// NewConsumer instantiates consumer
func NewConsumer(ep EventProcessor, q workqueue.RateLimitingInterface) *consumer {
	return &consumer{
		processor:             ep,
		queue:                 q,
		conciliationFrequency: defaultConciliationFrequency,
	}
}

// Run begins watching and syncing.
func (c *consumer) Run(stopCh chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.processor.Run(stopCh)

	go wait.Until(c.runWorker, c.conciliationFrequency, stopCh)

	<-stopCh
}

// runWorker executes the loop to process new items added to the queue
func (c *consumer) runWorker() {
	for c.processNextItem() {
	}
}

func (c *consumer) processNextItem() bool {
	ev, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(ev)

	e, ok := ev.(Event)
	if !ok {
		log.Error("nil event on processor")
		return true
	}

	err := c.processor.Handle(context.Background(), e)
	if err != nil {
		c.handleError(e, err)
		log.Errorf("event handled with key %q handled error: %v", e.GetKey(), err)
		return true
	}

	c.queue.Forget(e)
	return true
}

func (c *consumer) handleError(key interface{}, err error) {
	if c.queue.NumRequeues(key) < maxRequeue {
		log.Errorf("Error syncing pod %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	log.Errorf("event handled with key %v out of the queue: %v", key, err)

}

// Controller defines resource consumer runner
type Controller interface {
	Run(chan struct{})
}

// Build default Controller witch label selector option
func Build(lw ListWatcher, rh ResourceHandler, watchLabel ...string) Controller {
	eventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	eventHandler := NewResourceEventHandler(eventQueue)
	if len(watchLabel) > 0 {
		eventHandler = NewLabelSelectorMiddleware(watchLabel[0], eventHandler)
	}

	p := NewEventProcessor(lw, eventHandler, rh)
	return NewConsumer(p, eventQueue)

}
