package main

import "time"

// ScheduleMgr maintains a description of what the scheduler
// is be able to do
type ScheduleMgr interface {
	CreateOnetime(Arguments, GroupMgr, messageResponse)
}

// ScheduleMap should be the thing that holds the information
// for all the schedules produced
type ScheduleMap map[string][]*Schedule

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
}

// CreateOnetime schedules a message to be sent out once in the future
func (sm ScheduleMap) CreateOnetime(args Arguments, Groups GroupMgr, msgObj messageResponse) {
	schedule := new(Schedule)

	schedule.SessKey = msgObj.Room.GID + ":" + msgObj.Message.Sender.GID
	schedule.IsRecurring = false
	schedule.ExecuteOn, _ = time.Parse(time.RFC3339, args["dateTime"])
	schedule.UpdatedOn = time.Now()
	schedule.GroupID = Groups.GetGroup(args["groupName"]).ID // TODO Setup relationship
	schedule.MessageLabel = args["label"]
	schedule.MessageText = args["message"]

	sm[msgObj.Room.GID] = append(sm[msgObj.Room.GID], schedule)
}
