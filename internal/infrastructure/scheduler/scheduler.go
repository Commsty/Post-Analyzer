package scheduler

import (
	"errors"
	"fmt"
	"post-analyzer/internal/domain/entity"
	"strings"

	"github.com/robfig/cron/v3"
)

var (
	ErrSchedulingEvent = errors.New("sheduling failed")
)

type Scheduler interface {
	ScheduleEvent(sub *entity.Subscription, job func()) (int, error)
}

type scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *scheduler {
	return &scheduler{
		cron: cron.New(),
	}
}

func (s scheduler) Start() {
	s.cron.Start()
}

func (s scheduler) Stop() {
	s.cron.Stop()
}

func (s scheduler) ScheduleEvent(sub *entity.Subscription, job func()) (int, error) {

	parts := strings.Split(sub.SendingTime, ":")
	cronSpec := fmt.Sprintf("%s %s * * *", parts[1], parts[0])

	schedID, err := s.cron.AddFunc(cronSpec, job)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrSchedulingEvent, err)
	}

	return int(schedID), nil
}
