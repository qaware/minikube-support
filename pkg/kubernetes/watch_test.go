package kubernetes

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
)

func TestNewWatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name            string
		handler         WatchHandler
		options         *metav1.ListOptions
		resourceVersion string
		want            *Watcher
		wantErr         bool
	}{
		{
			"ok",
			fake.NewMockWatchHandler(ctrl),
			&metav1.ListOptions{},
			"1",
			&Watcher{"1", &metav1.ListOptions{}, fake.NewMockWatchHandler(ctrl), nil},
			false,
		},
		{
			"options nil",
			fake.NewMockWatchHandler(ctrl),
			nil,
			"1",
			&Watcher{"1", &metav1.ListOptions{}, fake.NewMockWatchHandler(ctrl), nil},
			false,
		},
		{
			"handler nil",
			nil,
			&metav1.ListOptions{},
			"1",
			nil,
			true,
		},
		{
			"resource empty",
			fake.NewMockWatchHandler(ctrl),
			&metav1.ListOptions{},
			"",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWatcher(tt.handler, tt.options, tt.resourceVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWatcher() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatcher_run(t *testing.T) {
	obj := &unstructured.Unstructured{}
	obj.SetResourceVersion("2")

	handler := &FakeWatchHandler{waiter: &sync.WaitGroup{}}
	handler.waiter.Add(1)
	watcher, _ := NewWatcher(handler, nil, "1")

	watcher.Start()
	handler.waiter.Wait()

	handler.watch.Add(obj)
	handler.watch.Modify(obj)

	handler.waiter.Add(1)
	handler.watch.Error(nil)
	handler.waiter.Wait()

	handler.watch.Delete(obj)
	handler.AddPostError(errors.New(""))
	handler.watch.Add(obj)

	handler.AddError(errors.New(""))
	handler.watch.Add(obj)
	handler.watch.Modify(obj)
	handler.watch.Delete(obj)
	time.Sleep(10 * time.Millisecond)
	watcher.Stop()

	expectedMethods := []string{"PreWatch",
		"AddedEvent",
		"PostEvent",
		"UpdatedEvent",
		"PostEvent",
		"PreWatch",
		"DeletedEvent",
		"PostEvent",
		"AddedEvent",
		"PostEvent",
		"AddedEvent",
		"UpdatedEvent",
		"DeletedEvent"}
	if !reflect.DeepEqual(expectedMethods, handler.GetExecutedFunc()) {
		t.Errorf("Wrong execution order:\nWant: %s\nGot:  %s", expectedMethods, handler.executedFunc)
	}
}

type FakeWatchHandler struct {
	executedFunc []string
	watch        *watch.FakeWatcher
	error        error
	postError    error
	lock         sync.Mutex
	waiter       *sync.WaitGroup
}

func (h *FakeWatchHandler) AddError(err error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.error = err
}
func (h *FakeWatchHandler) AddPostError(err error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.postError = err
}

func (h *FakeWatchHandler) GetExecutedFunc() []string {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.executedFunc
}

func (h *FakeWatchHandler) PreWatch(_ metav1.ListOptions) (watch.Interface, error) {
	h.lock.Lock()
	defer func() {
		h.lock.Unlock()
		h.waiter.Done()
	}()
	h.executedFunc = append(h.executedFunc, "PreWatch")
	h.watch = watch.NewFake()
	return h.watch, h.error
}

func (h *FakeWatchHandler) AddedEvent(_ runtime.Object) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.executedFunc = append(h.executedFunc, "AddedEvent")
	return h.error
}

func (h *FakeWatchHandler) UpdatedEvent(_ runtime.Object) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.executedFunc = append(h.executedFunc, "UpdatedEvent")
	return h.error
}

func (h *FakeWatchHandler) DeletedEvent(_ runtime.Object) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.executedFunc = append(h.executedFunc, "DeletedEvent")
	return h.error
}

func (h *FakeWatchHandler) PostEvent() error {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.executedFunc = append(h.executedFunc, "PostEvent")
	return h.postError
}
