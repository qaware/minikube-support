package kubernetes

//go:generate mockgen -destination=fake/watchMocks.go -package=fake -source=watch.go

import (
	"errors"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// Watcher is a small helper to handle to watch changes on kubernetes objects. It
// reduces the need of error handling in the actual implementation.
type Watcher struct {
	// Resource Version is the minimum version to start the watch on.
	ResourceVersion string
	options         *metav1.ListOptions
	handler         WatchHandler
	watch           watch.Interface
}

// WatchHandler is the interface that must be implemented by the application code
// which requires watching on kubernetes resources.
type WatchHandler interface {
	// PreWatch is executed before the watch loop starts and must return the
	// watch.Interface instance to watch for the events.
	PreWatch(options metav1.ListOptions) (watch.Interface, error)
	// AddedEvent handles the event when a new resource was added.
	AddedEvent(obj runtime.Object) error
	// UpdatedEvent handles the event when a resource was updated.
	UpdatedEvent(obj runtime.Object) error
	// DeletedEvent handles the event when a resource was deleted.
	DeletedEvent(obj runtime.Object) error
	// PostEvent is executed after every event if the event handling were executed
	// without an error.
	PostEvent() error
}

// NewWatcher initializes a new Watcher instance and ensures the consistency
// of the given parameters.
func NewWatcher(handler WatchHandler, options *metav1.ListOptions, resourceVersion string) (*Watcher, error) {
	if handler == nil {
		return nil, errors.New("watch handler is nil")
	}
	if options == nil {
		options = &metav1.ListOptions{}
	}
	if resourceVersion == "" {
		return nil, errors.New("resource version is empty. watcher can not be started without an resource version")
	}
	return &Watcher{
		ResourceVersion: resourceVersion,
		options:         options,
		handler:         handler,
	}, nil
}

// Start starts the watcher. It will immediately return.
func (h *Watcher) Start() {
	go h.run()
}

// Stop ends the watcher loop.
func (h *Watcher) Stop() {
	if h.watch != nil {
		h.watch.Stop()
	}
}

func (h *Watcher) run() {
	restartWatch := true
	for restartWatch {
		restartWatch = h.watcher()
	}
}

func (h *Watcher) watcher() bool {
	options := h.options.DeepCopy()
	options.ResourceVersion = h.ResourceVersion

	w, e := h.handler.PreWatch(*options)
	if e != nil {
		logrus.Errorf("Can not start watch: %s", e)
		return false
	}
	h.watch = w
	accessor := meta.NewAccessor()

	for event := range w.ResultChan() {
		var e error
		switch event.Type {
		case watch.Added:
			e = h.handler.AddedEvent(event.Object)
		case watch.Modified:
			e = h.handler.UpdatedEvent(event.Object)
		case watch.Deleted:
			e = h.handler.DeletedEvent(event.Object)
		case watch.Error:
			h.watch.Stop()
			logrus.Infof("Got Error Event: %v Restart watch now.", event.Object)
			return true
		default:
			logrus.Infof("Received unhandled event %s for object %v", event.Type, event.Object)
		}

		if e != nil {
			logrus.Warnf("Can not handle %s event: %s", event.Type, e)
			continue
		}

		resourceVersion, e := accessor.ResourceVersion(event.Object)
		if e != nil {
			logrus.Debugf("can not extract resource version: %s", e)
		} else {
			h.ResourceVersion = resourceVersion
		}

		e = h.handler.PostEvent()
		if e != nil {
			logrus.Info("Unable to handle post event function")
		}
	}
	return false
}
