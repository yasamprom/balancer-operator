package usecases

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/yasamprom/balancer-operator/internal/model"
	slicer "github.com/yasamprom/balancer-operator/internal/repo/clients/slicer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Config ...
type Config struct {
	Watcher  watch.Interface
	Triggers model.Triggers
	Slicer   *slicer.Client
}

// Watcher ...
type Watcher struct {
	w           watch.Interface
	triggers    model.Triggers
	stuckEvents model.UpdateNodes
	slicer      *slicer.Client
}

// New creates new watcher
func New(c Config) *Watcher {
	return &Watcher{
		w:        c.Watcher,
		triggers: c.Triggers,
		slicer:   c.Slicer,
	}
}

// StartWatching runs manager for handling events on pods
func (w *Watcher) StartWatching(ctx context.Context) error {
	go func() error {

		// configure ticker for sending events by chunks
		ticker := time.NewTicker(500 * time.Millisecond)
		var events model.UpdateNodes

		for {
			select {
			case <-ctx.Done():
				return nil

			case event := <-w.w.ResultChan():
				if !w.shouldProcess(event) {
					continue
				}

				if event.Type == watch.Added {
					pod := event.Object.(*corev1.Pod)
					log.Printf("registered pod: %s, %s\n", pod.Name, pod.Status.PodIP)
					events.New.Hosts = append(events.New.Hosts, model.Address{
						Host: pod.Status.PodIP,
					})
				}
				if event.Type == watch.Deleted {
					pod := event.Object.(*corev1.Pod)
					log.Printf("deleted pod: %s, %s\n", pod.Name, pod.Status.PodIP)
					events.Disconnected.Hosts = append(events.Disconnected.Hosts, model.Address{
						Host: pod.Status.PodIP,
					})
				}
				if event.Type == watch.Error {
					pod := event.Object.(*corev1.Pod)
					log.Printf("error on pod: %s, %s\n", pod.Name, pod.Status.PodIP)
					// to be handled
				}

			// send events chunk if it is not empty
			case <-ticker.C:
				if events.ContainsEvents() {
					err := w.slicer.NotifyEvents(ctx, events)
					if err != nil {
						log.Printf("watcher couldn't send events: %v", err)
						w.stuckEvents.New.Hosts = append(w.stuckEvents.New.Hosts, events.New.Hosts...)
						w.stuckEvents.Disconnected.Hosts = append(w.stuckEvents.Disconnected.Hosts, events.Disconnected.Hosts...)
					}
					// clear events
					events = model.UpdateNodes{}
				}
			}
		}
	}()

	// send undelivered events
	stuckTicker := time.NewTicker(1000 * time.Millisecond)
	go func() {
		select {
		case <-stuckTicker.C:
			if w.stuckEvents.ContainsEvents() {
				err := w.slicer.NotifyEvents(ctx, w.stuckEvents)
				if err == nil {
					// clear events if sent successfully
					w.stuckEvents = model.UpdateNodes{}
				} else {
					log.Printf("watcher couldn't send stuck events: %v", err)
				}
			}
		case <-ctx.Done():
			return
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
