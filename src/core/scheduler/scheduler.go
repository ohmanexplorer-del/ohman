package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron   *cron.Cron
	tasks  map[string]cron.EntryID
	ctx    context.Context
	cancel context.CancelFunc
}

type Task struct {
	Name     string
	Schedule string
	Func     func() error
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:   cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger))),
		tasks:  make(map[string]cron.EntryID),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) AddTask(task Task) error {
	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("task %s already exists", task.Name)
	}

	wrappedFunc := func() {
		select {
		case <-s.ctx.Done():
			return
		default:
			log.Printf("Executing task: %s", task.Name)
			if err := task.Func(); err != nil {
				log.Printf("Task %s failed: %v", task.Name, err)
			} else {
				log.Printf("Task %s completed successfully", task.Name)
			}
		}
	}

	id, err := s.cron.AddFunc(task.Schedule, wrappedFunc)
	if err != nil {
		return fmt.Errorf("failed to add task %s: %w", task.Name, err)
	}

	s.tasks[task.Name] = id
	log.Printf("Task %s scheduled with cron: %s", task.Name, task.Schedule)
	return nil
}

func (s *Scheduler) RemoveTask(name string) error {
	id, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	s.cron.Remove(id)
	delete(s.tasks, name)
	log.Printf("Task %s removed", name)
	return nil
}

func (s *Scheduler) Start() {
	log.Println("Starting scheduler...")
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cancel()
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) IsTaskRunning(name string) bool {
	id, exists := s.tasks[name]
	if !exists {
		return false
	}
	entry := s.cron.Entry(id)
	return entry.ID != cron.EntryID(0)
}
