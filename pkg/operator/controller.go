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

type EventProcessor interface {
	Run(stopCh chan struct{})
	Handle(ctx context.Context, ev Event) error
}

type controller struct {
	processor             EventProcessor
	queue                 workqueue.RateLimitingInterface
	conciliationFrequency time.Duration
}

func NewController(ep EventProcessor, q workqueue.RateLimitingInterface) *controller {
	return &controller{
		processor:             ep,
		queue:                 q,
		conciliationFrequency: defaultConciliationFrequency,
	}
}

// Run begins watching and syncing.
func (c *controller) Run(stopCh chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.processor.Run(stopCh)

	go wait.Until(c.runWorker, c.conciliationFrequency, stopCh)

	<-stopCh
}

// runWorker executes the loop to process new items added to the queue
func (c *controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
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

func (c *controller) handleError(key interface{}, err error) {
	if c.queue.NumRequeues(key) < maxRequeue {
		log.Errorf("Error syncing pod %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	log.Errorf("event handled with key %v out of the queue: %v", key, err)

}
