package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	chat "google.golang.org/api/chat/v1"
)

// ScheduleMgr maintains a description of what the scheduler
// is be able to do
type ScheduleMgr interface {
	CreateOnetime(Arguments, GroupMgr, messageResponse) string
	CreateRecurring(Arguments, GroupMgr, messageResponse) string
	Remove(Arguments, messageResponse) string
	List(messageResponse) string
}

// ScheduleMap should be the thing that holds the information
// for all the schedules produced
type ScheduleMap map[string]*Schedule

// Schedule contains information required to schedule a
// message
type Schedule struct {
	ID           uint      `gorm:"primary_key;not null;unique" yaml:"-"`
	SessKey      string    `gorm:"not null" yaml:"-"` // room_id:user_id
	Creator      string    `gorm:"not null" yaml:"creator"`
	IsRecurring  bool      `gorm:"not null" yaml:"recurring"`
	DayKey       string    `yaml:"-"`
	CreatedOn    time.Time `gorm:"not null" yaml:"createdOn"`
	ExecuteOn    time.Time `gorm:"not null" yaml:"sendOn"`
	UpdatedOn    time.Time `yaml:"updatedOn,omitempty"`
	CompletedOn  time.Time `yaml:"completedOn,omitempty"`
	GroupID      uint      `gorm:"not null" yaml:"-"`
	ThreadKey    string    `gorm:"not null" yaml:"-"`
	MessageLabel string    `gorm:"not null" yaml:"label"`
	MessageText  string    `gorm:"not null" yaml:"message"`
	IsFinished   bool      `gorm:"not null;default:false" yaml:"-"`
	timer        *time.Timer
}

// CreateOnetime schedules a message to be sent out once in the future
func (sm ScheduleMap) CreateOnetime(args Arguments, Groups GroupMgr, msgObj messageResponse) string {
	var schedule *Schedule

	schedKey := msgObj.Room.GID + ":" + args["label"]

	if sm.hasSchedule(schedKey) {
		schedule = sm.getSchedule(schedKey)
		schedule.UpdatedOn = time.Now()
	} else {
		schedule = new(Schedule)
		schedule.CreatedOn = time.Now()
	}

	schedule.SessKey = msgObj.Room.GID + ":" + msgObj.Message.Sender.GID
	schedule.Creator = msgObj.Message.Sender.Name
	schedule.IsRecurring = false
	schedule.ExecuteOn, _ = time.Parse(time.RFC3339, args["dateTime"])
	schedule.GroupID = Groups.GetGroup(args["groupName"]).Model.ID // TODO Setup relationship
	schedule.ThreadKey = msgObj.Message.Thread.Name
	schedule.MessageLabel = args["label"]
	schedule.MessageText = args["message"]

	schedule.StartTimer()

	sm[schedKey] = schedule

	go Logger.SaveSchedule(schedule)

	return fmt.Sprintf("Scheduled onetime message %q for group %q to be sent on %q",
		schedule.MessageLabel,
		args["groupName"],
		schedule.ExecuteOn.Format("Monday, 2 January 2006 3:04 PM MST"),
	)
}

// CreateRecurring schedules a message to be sent out weekly in the future
func (sm ScheduleMap) CreateRecurring(args Arguments, Groups GroupMgr, msgObj messageResponse) string {
	var schedule *Schedule

	schedKey := msgObj.Room.GID + ":" + args["label"]

	if sm.hasSchedule(schedKey) {
		schedule = sm.getSchedule(schedKey)
		schedule.UpdatedOn = time.Now()
	} else {
		schedule = new(Schedule)
		schedule.CreatedOn = time.Now()
	}

	schedule.SessKey = msgObj.Room.GID + ":" + msgObj.Message.Sender.GID
	schedule.Creator = msgObj.Message.Sender.Name
	schedule.IsRecurring = true
	schedule.ExecuteOn, _ = time.Parse(time.RFC3339, args["dateTime"])
	schedule.GroupID = Groups.GetGroup(args["groupName"]).Model.ID
	schedule.ThreadKey = msgObj.Message.Thread.Name
	schedule.MessageLabel = args["label"]
	schedule.MessageText = args["message"]
	schedule.IsFinished = false

	schedule.StartTimer()

	sm[schedKey] = schedule

	go Logger.SaveSchedule(schedule)

	return fmt.Sprintf("Scheduled recurring message %q for group %q to be sent on %s.\n",
		schedule.MessageLabel,
		args["groupName"],
		fmt.Sprintf(
			"%s's @ %s, starting on %s",
			schedule.ExecuteOn.Weekday(),
			schedule.ExecuteOn.Format(time.Kitchen),
			schedule.ExecuteOn.Format("Monday, January 2, 2006"),
		),
	)
}

// List returns a yaml list of the schedules specific to the
// room the request came from
//
// TODO: Fix how timestamps display when marhsalling to yaml.
// Currently, it looks like there is no way to change how a
// time.Time object displayes when using the go-yaml/yaml lib.
// Likely the solution would be to choose a different library.
func (sm ScheduleMap) List(msgObj messageResponse) string {
	curRoomID := msgObj.Room.GID

	var schedList string

	for schedKey, schedule := range sm {
		schedRoomID := strings.Split(schedKey, ":")[0]

		if !schedule.IsFinished && schedRoomID == curRoomID {
			data, err := yaml.Marshal(schedule)
			checkError(err)

			schedList += "\n" + string(data)
		}
	}

	if schedList == "" {
		return fmt.Sprintf("There are upcoming scheduled messages setup for this room as of yet.")
	}

	return fmt.Sprintf("Here are a list of the scheduled messages for this room ```%s```", schedList)
}

// Remove marks the specified schedule as finished, saves these changes to
// the database, then deletes the key for the schedule.
func (sm ScheduleMap) Remove(args Arguments, msgObj messageResponse) string {
	label := args["label"]
	roomID := msgObj.Room.GID

	schedKey := roomID + ":" + label

	if sm.hasSchedule(schedKey) {
		schedule := sm.getSchedule(schedKey)
		schedule.timer.Stop()
		schedule.IsFinished = true

		Logger.SaveSchedule(schedule)

		delete(sm, schedKey)

		return fmt.Sprintf("Removed %q from scheduler.", label)
	}

	return fmt.Sprintf("Message %q not found to be removed.", label)
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
func (sm ScheduleMap) getSchedule(schedKey string) *Schedule {
	return sm[schedKey]
}

// HasSchedule checks to see if the schedule being called
// already exists
func (sm ScheduleMap) hasSchedule(schedKey string) bool {
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

		if !s.IsFinished && time.Now().After(s.ExecuteOn) {
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

	s.complete()
}

func (s *Schedule) complete() {
	s.CompletedOn = time.Now()

	if s.IsRecurring {
		// Since the scheduled messages are all weekly, once they've
		// completed a run, add 7 days (168 hours)
		s.ExecuteOn.Add(time.Hour * 168)
		s.StartTimer()
	} else {
		s.IsFinished = true
	}

	Logger.SaveSchedule(s)
}
