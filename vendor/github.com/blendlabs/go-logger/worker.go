package logger

const (
	// DefaultWorkerQueueDepth is the default depth per listener to queue work.
	DefaultWorkerQueueDepth = 1 << 20
)

// NewWorker returns a new worker.
func NewWorker(parent *Logger, listener Listener) *Worker {
	return &Worker{
		Parent:   parent,
		Listener: listener,
		Work:     make(chan Event, DefaultWorkerQueueDepth),
		Abort:    make(chan bool),
		Aborted:  make(chan bool),
	}
}

// Worker is an agent that processes a listener.
type Worker struct {
	Parent    *Logger
	Listener  Listener
	Abort     chan bool
	Aborted   chan bool
	Drained   chan bool
	Work      chan Event
	IsRunning bool
}

// Start starts the worker.
func (w *Worker) Start() {
	go w.ProcessLoop()
}

// ProcessLoop is the for/select loop.
func (w *Worker) ProcessLoop() {
	var e Event
	w.IsRunning = true
	for w.IsRunning {
		select {
		case e = <-w.Work:
			w.Process(e)
		case <-w.Abort:
			w.IsRunning = false
			w.Aborted <- true
			return
		}
	}
}

// Process calls the listener for an event.
func (w *Worker) Process(e Event) {
	defer func() {
		if r := recover(); r != nil {
			if w.Parent != nil {
				w.Parent.SyncFatalf("%v", r)
			}
		}
	}()
	w.Listener(e)
}

// Stop stops the worker.
func (w *Worker) Stop() {
	if !w.IsRunning {
		return
	}
	w.Abort <- true
	<-w.Aborted
}

// Drain stops the worker and synchronously processes any remaining work.
func (w *Worker) Drain() {
	w.Stop()
	for len(w.Work) > 0 {
		w.Process(<-w.Work)
	}
}

// Close closes the worker.
func (w *Worker) Close() error {
	w.Stop()
	close(w.Work)
	close(w.Abort)
	close(w.Aborted)
	w.Work = nil
	w.Abort = nil
	w.Aborted = nil
	return nil
}
