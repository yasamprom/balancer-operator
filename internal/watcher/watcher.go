package usecases

import (
	"fmt"
	"strings"

	"github.com/yasamprom/balancer-operator/internal/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Config ...
type Config struct {
	Watcher  watch.Interface
	Triggers model.Triggers
}

// Watcher ...
type Watcher struct {
	w        watch.Interface
	triggers model.Triggers
}

// New creates new watcher
func New(c Config) *Watcher {
	return &Watcher{
		w:        c.Watcher,
		triggers: c.Triggers,
	}
}

func (w *Watcher) StartWatching() error {
	go func() error {
		for {
			select {
			case event := <-w.w.ResultChan():
				if !w.shouldProcess(event) {
					continue
				}
				if event.Type == watch.Added {
					pod := event.Object.(*corev1.Pod)
					fmt.Printf("registered pod: %s, %s\n", pod.Name, pod.Status.PodIP)
				}
				if event.Type == watch.Deleted {
					pod := event.Object.(*corev1.Pod)
					fmt.Printf("deleted pod: %s, %s\n", pod.Name, pod.Status.PodIP)
				}
				if event.Type == watch.Error {
					pod := event.Object.(*corev1.Pod)
					fmt.Printf("error on pod: %s, %s\n", pod.Name, pod.Status.PodIP)
				}

			default:
				continue
			}
		}
	}()
	return nil
}

func (w *Watcher) shouldProcess(in watch.Event) bool {
	pod := in.Object.(*corev1.Pod)
	name := pod.Name
	if nameMatches(w.triggers.Names, name) {
		return true
	}
	// other triggers
	return false
}

func nameMatches(prefs []string, val string) bool {
	for _, prefix := range prefs {
		if strings.HasPrefix(val, prefix) {
			return true
		}
	}
	return false
}
