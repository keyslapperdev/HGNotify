package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	chat "google.golang.org/api/chat/v1"
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
	Creator      string `gorm:"not null"`
	IsRecurring  bool   `gorm:"not null"`
	DayKey       string
	CreatedOn    time.Time `gorm:"not null"`
	ExecuteOn    time.Time `gorm:"not null"`
	UpdatedOn    time.Time
	CompletedOn  time.Time
	GroupID      uint   `gorm:"not null"`
	ThreadKey    string `gorm:"not null"`
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
	schedule.Creator = msgObj.Message.Sender.Name
	schedule.IsRecurring = false
	schedule.CreatedOn = time.Now()
	schedule.ExecuteOn, _ = time.Parse(time.RFC3339, args["dateTime"])
	schedule.GroupID = Groups.GetGroup(args["groupName"]).Model.ID // TODO Setup relationship
	schedule.ThreadKey = msgObj.Message.Thread.Name
	schedule.MessageLabel = args["label"]
	schedule.MessageText = args["message"]

	schedule.StartTimer()

	sm[schedKey] = schedule

	go Logger.SaveScheduledEvent(schedule)

	return fmt.Sprintf("Scheduled onetime message %q for group %q to be sent on %q",
		schedule.MessageLabel,
		args["groupName"],
		schedule.ExecuteOn.Format("Monday, 2 January 2006 3:04 PM CDT"),
	)
}

// GetLabels returns a list of the rooms schedules have been created for
func (sm ScheduleMap) GetLabels() []string {
	var labels []string

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
	// As of right now, because the logic to generate the "Notify"
	// message soley lives within a method for GroupMap and requires
	// a messageResponse object to create, to get the message I have
	// to make an instance of GroupMap and messageResponse to retrieve
	// the message. Even for MVP this is grody, but a baby's gotta do
	// what a baby's gotta do.
	groups := make(GroupMap)
	group := Logger.GetGroupByID(s.GroupID)
	groups[group.Name] = group

	msgObj := messageResponse{}
	// Mimicking how the message would normally look
	msgObj.Message.Text = BotName + " " + group.Name + " " + s.MessageText
	msgObj.Message.Sender.Name = s.Creator

	msg := groups.Notify(group.Name, msgObj)

	chatService := getChatService(getChatClient())
	msgService := chat.NewSpacesMessagesService(chatService)

	room := strings.Split(s.SessKey, ":")[0]

	_, err := msgService.Create(room, &chat.Message{
		Text: msg,
		Thread: &chat.Thread{
			Name: s.ThreadKey,
		},
	}).Do()
	if err != nil {
		log.Fatal("Error sending scheduled notification: " + err.Error())
	}

	s.CompletedOn = time.Now()
}
