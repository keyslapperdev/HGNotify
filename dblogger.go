package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//DBLogger struct just isa pointer to a gorm.DB object, it just holds
//the DB object so that I can throw methods on it.
type DBLogger struct {
	*gorm.DB
	isActive bool
}

//NotifyLog defines a database schema for the message logger.
type NotifyLog struct {
	MessageID uint      `gorm:"primary_key;not null;unique"`
	TimeSent  time.Time `gorm:"not null"`
	Sender    string    `gorm:"not null"`
	Message   string    `gorm:"type:varchar(4000);not null"`
}

//startDBLogger is used to intialize the logger
func startDBLogger(conf DBConfig) DBLogger {
	//I format everything with gofmt. Because of the colon after DBUser
	//makes gofmt think the lower lines are apart of switch case statement.
	//Hence the weird indention. I'd rather everything be formatted as gofmt
	//dictates than not. /shrug
	db, err := gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@(%s)/%s?%s",
		conf.DBUser,
		conf.DBPass,
		conf.DBHost,
		conf.DBName,
		"charset=utf8mb4&parseTime=True",
	))

	checkError(err)

	return DBLogger{db, true}
}

//SetupTables method is used when the file first starts up to
//ensure all databases are created and updated as they should be.
//If everything is fine, this is only ran once, ever.
func (db *DBLogger) SetupTables() {
	if !db.isActive {
		return
	}
	db.AutoMigrate(&Group{}, &Member{}, &NotifyLog{}, &Schedule{})
	db.Model(&Member{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "RESTRICT")
}

//SaveCreatedGroup method is used to update the database whenever
//a new group is created
func (db *DBLogger) SaveCreatedGroup(group *Group) {
	if !db.isActive {
		return
	}
	db.Create(group)
}

//DisbandGroup method will delete a group's entry from the database,
//along with all the associated users.
func (db *DBLogger) DisbandGroup(group *Group) {
	if !db.isActive {
		return
	}
	db.Unscoped().Delete(group)
}

//UpdatePrivacyDB method toggles the privacy settings for the specified
//group. It's a bit different because when the restriction is removed, the
//values entered into the database are "zero value", so gorm ignores them.
//To get them to set the zero value I have to be specific with the query.
func (db *DBLogger) UpdatePrivacyDB(group *Group) {
	if !db.isActive {
		return
	}
	db.Model(group).Select("is_private").Update("IsPrivate", group.IsPrivate)
	db.Model(group).Select("privacy_room_id").Update("PrivacyRoomID", group.PrivacyRoomID)
}

//SaveMemberAddition method adds a member to the associated group
func (db *DBLogger) SaveMemberAddition(group *Group) {
	if !db.isActive {
		return
	}
	db.Model(group).Update(group)
}

//SaveMemberRemoval method marks the assocaited memeber as removed from
//the group.
func (db *DBLogger) SaveMemberRemoval(group *Group, members []Member) {
	if !db.isActive {
		return
	}
	for _, member := range members {
		db.Model(group).Delete(member)
	}
}

//GetGroupsFromDB method syncs the database groups to the in-memory group list
//this is ran when the program starts up.
func (db *DBLogger) GetGroupsFromDB(groupMap GroupMap) {
	if !db.isActive {
		return
	}
	var foundGroups []*Group
	db.Find(&foundGroups)

	for _, group := range foundGroups {
		var members []Member
		db.Model(&group).Related(&members)

		group.Members = members

		saveName := strings.ToLower(group.Name)
		groupMap[saveName] = group
	}
}

//GetGroupByID pulls a group from the databse given the ID.
//I'm aware that this kinda defies the "Logger" nature of this
//class, but as of now this is something I need to keep going.
//Major refactors are going to come after I'm done with the
//schedule features
func (db *DBLogger) GetGroupByID(groupID uint) *Group {
	if !db.isActive {
		return nil
	}

	groupSearchTmpl := new(Group)
	groupSearchTmpl.Model.ID = groupID

	group := new(Group)

	db.Model(Group{}).Where(groupSearchTmpl).Find(group)

	memberSearchTmpl := new(Member)
	memberSearchTmpl.GroupID = group.Model.ID

	members := make([]Member, 0)

	db.Model(Member{}).Where(memberSearchTmpl).Find(&members)

	group.Members = members

	return group
}

//SyncAllGroups method syncs the database groups to the in-memory group list
//during runtime. Just in case.
func (db *DBLogger) SyncAllGroups(groups GroupMap) {
	if !db.isActive {
		return
	}

	//TODO: Create workergroup to have about 10 groups to be updated at a time.
	var wg sync.WaitGroup

	//A go routine is spawned for each group to sync. It's a bit intensive on memory
	//but speeds up the process significatly.
	for _, group := range groups {
		wg.Add(1)
		go func(group *Group, wg *sync.WaitGroup) {
			Logger.SyncGroup(group)
			wg.Done()
		}(group, &wg)
	}

	wg.Wait()
}

//SyncGroup method syncs the members in an in-memory group to the database entries
//this can be done during runtime. Just in case.
func (db *DBLogger) SyncGroup(group *Group) {
	if !db.isActive {
		return
	}

	var members []Member
	db.Model(&group).Related(&members)

	group.Members = members
}

//CreateLogEntry method logs usage of the bot to the database.
func (db *DBLogger) CreateLogEntry(msgObj messageResponse) {
	if !db.isActive {
		return
	}

	sentAt, _ := time.Parse(time.RFC3339Nano, msgObj.Time)

	//This is the result of a weird bug with the way time.Time.In returns/modifies
	//the Time object. This is gross and I apologize to anyoene who ever sees this.
	sentAt = sentAt.Add(-time.Hour * 5)

	entry := &NotifyLog{
		TimeSent: sentAt,
		Sender:   msgObj.Message.Sender.Name,
		Message:  msgObj.Message.Text,
	}

	db.Create(entry)
}

// SaveScheduledEvent saves the event to the database
func (db *DBLogger) SaveScheduledEvent(schedule *Schedule) {
	if !db.isActive {
		return
	}

	if schedule.ID != 0 {
		db.Save(schedule)
		return
	}

	db.Create(schedule)
}

// GetSchedulesFromDB Grabbing all of the schedules from the
// db to be consumed at app startup
func (db *DBLogger) GetSchedulesFromDB(sMap ScheduleMap) {
	if !db.isActive {
		return
	}

	var schedules []Schedule
	db.Find(&schedules)

	for _, schedule := range schedules {
		if !schedule.CompletedOn.IsZero() && !schedule.IsRecurring {
			continue
		}

		room := strings.Split(schedule.SessKey, ":")[0]
		label := schedule.MessageLabel

		schedule.StartTimer()

		sMap[room+":"+label] = &schedule
	}
}

//Active is the setter method for the activity of the db logger
func (db *DBLogger) Active(status bool) {
	db.isActive = status
}
