package main

import (
	"fmt"
	"time"
)

// ScheduleMgr maintains a description of what the scheduler
// is be able to do
type ScheduleMgr interface {
	CreateOnetime(Arguments, GroupMgr, messageResponse) string
}

// ScheduleMap should be the thing that holds the information
// for all the schedules produced
type ScheduleMap map[string]*Schedule

// Schedule contains information required to schedule a
// message
type Schedule struct {
	ID           uint   `gorm:"primary_key;not null;unique"`
	SessKey      string `gorm:"not null"` // room_id:user_id
	IsRecurring  bool   `gorm:"not null"`
	DayKey       string
	ExecuteOn    time.Time `gorm:"not null"`
	UpdatedOn    time.Time `gorm:"not null"`
	CompletedOn  time.Time
	GroupID      uint   `gorm:"not null"`
	MessageLabel string `gorm:"not null"`
	MessageText  string `gorm:"not null"`
	timer        *time.Timer
}

// CreateOnetime schedules a message to be sent out once in the future
func (sm ScheduleMap) CreateOnetime(args Arguments, Groups GroupMgr, msgObj messageResponse) string {
	var schedule *Schedule

	schedKey := msgObj.Room.GID + ":" + args["label"]

	if sm.HasSchedule(schedKey) {
		schedule = sm.GetSchedule(schedKey)
		schedule.UpdatedOn = time.Now()
	} else {
		schedule = new(Schedule)
	}

	schedule.SessKey = msgObj.Room.GID + ":" + msgObj.Message.Sender.GID
	schedule.IsRecurring = false
	schedule.ExecuteOn, _ = time.Parse(time.RFC3339, args["dateTime"])
	schedule.UpdatedOn = time.Now()
	schedule.GroupID = Groups.GetGroup(args["groupName"]).ID // TODO Setup relationship
	schedule.MessageLabel = args["label"]
	schedule.MessageText = args["message"]

	schedule.StartTimer()

	sm[schedKey] = schedule

	return fmt.Sprintf("Scheduled onetime message %q for group %q to be sent on %q",
		schedule.MessageLabel,
		args["groupName"],
		schedule.ExecuteOn.Format(time.RFC850),
	)
}

// GetLabels returns a list of the rooms schedules have been created for
func (sm ScheduleMap) GetLabels() []string {
	labels := make([]string, len(sm))

	for label := range sm {
		labels = append(labels, label)
	}

	return labels
}

// GetSchedule returns the schedule for the given label
func (sm ScheduleMap) GetSchedule(schedKey string) *Schedule {
	return sm[schedKey]
}

// HasSchedule checks to see if the schedule being called
// already exists
func (sm ScheduleMap) HasSchedule(schedKey string) bool {
	_, exists := sm[schedKey]
	return exists
}

// StartTimer begins the countdown until the message is sent or sends
// immediately if message is overdue
func (s *Schedule) StartTimer() {
	go func() {
		if s.timer != nil {
			s.timer.Stop()
		}

		if !s.IsRecurring &&
			time.Now().After(s.ExecuteOn) &&
			s.CompletedOn.IsZero() {
			s.Send()
			return
		}

		s.timer = time.AfterFunc(
			time.Until(s.ExecuteOn),
			func() { s.Send() },
		)
	}()
}

// Send will send out the message scheduled
func (s *Schedule) Send() {
	go func() {
		s.CompletedOn = time.Now()
	}()
}
