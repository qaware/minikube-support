package kubernetes

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/qaware/minikube-support/pkg/kubernetes/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"reflect"
	"testing"
	"time"
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

	handler := &FakeWatchHandler{}
	watcher, _ := NewWatcher(handler, nil, "1")
	watcher.Start()

	time.Sleep(10 * time.Millisecond)
	handler.watch.Add(obj)
	handler.watch.Modify(obj)
	handler.watch.Error(nil)

	time.Sleep(10 * time.Millisecond)
	handler.watch.Delete(obj)

	handler.postError = errors.New("")
	handler.watch.Add(obj)

	handler.error = errors.New("")

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
	if !reflect.DeepEqual(expectedMethods, handler.executedFunc) {
		t.Errorf("Wrong execution order:\nWant: %s\nGot:  %s", expectedMethods, handler.executedFunc)
	}
}

type FakeWatchHandler struct {
	executedFunc []string
	watch        *watch.FakeWatcher
	error        error
	postError    error
}

func (h *FakeWatchHandler) PreWatch(options metav1.ListOptions) (watch.Interface, error) {
	h.executedFunc = append(h.executedFunc, "PreWatch")
	h.watch = watch.NewFake()
	return h.watch, h.error
}

func (h *FakeWatchHandler) AddedEvent(obj runtime.Object) error {
	h.executedFunc = append(h.executedFunc, "AddedEvent")
	return h.error
}

func (h *FakeWatchHandler) UpdatedEvent(obj runtime.Object) error {
	h.executedFunc = append(h.executedFunc, "UpdatedEvent")
	return h.error
}

func (h *FakeWatchHandler) DeletedEvent(obj runtime.Object) error {
	h.executedFunc = append(h.executedFunc, "DeletedEvent")
	return h.error
}

func (h *FakeWatchHandler) PostEvent() error {
	h.executedFunc = append(h.executedFunc, "PostEvent")
	return h.postError
}
