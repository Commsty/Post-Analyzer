package scheduler

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) ScheduleJob(hour, minute int, job func()) (int, error) {

	cronSpec := fmt.Sprintf("%d %d * * *", minute, hour)
	schedID, err := s.cron.AddFunc(cronSpec, job)
	return int(schedID), err
}

func (s *Scheduler) ParseTimeString(timeString string) (int, int, error) {

	timeString = strings.TrimSpace(timeString)

	var hour, minute int
	n, err := fmt.Sscanf(timeString, "%d:%d", &hour, &minute)
	if err != nil || n != 2 {
		return -1, -1, fmt.Errorf("invalid time format. Expected format: HH:MM")
	}

	if hour < 0 || hour > 23 {
		return -1, -1, fmt.Errorf("hour value must be between 0 and 23")
	}

	if minute < 0 || minute > 59 {
		return -1, -1, fmt.Errorf("minute value must be between 0 and 59")
	}

	return hour, minute, nil
}
