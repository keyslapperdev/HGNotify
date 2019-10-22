package main

import (
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DBLogger struct {
	*gorm.DB
}

type NotifyLog struct {
	MessageID uint      `gorm:"primary_key;not null;unique"`
	TimeSent  time.Time `gorm:"not null"`
	Sender    string    `gorm:"not null"`
	Message   string    `gorm:"type:varchar(4000);not null"`
}

func startDBLogger(conf DBConfig) DBLogger {
	db, err := gorm.Open(
		conf.Driver,
		conf.DBUser+":"+
			conf.DBPass+"@/"+
			conf.DBName+"?"+
			conf.DBOpts)

	checkError(err)

	return DBLogger{db}
}

func (db *DBLogger) SetupTables() {
	db.AutoMigrate(&Group{}, &Member{}, &NotifyLog{})
	db.Model(&Member{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "RESTRICT")
}

func (db *DBLogger) SaveCreatedGroup(group *Group) {
	db.Create(group)
}

func (db *DBLogger) DisbandGroup(group *Group) {
	db.Unscoped().Delete(group)
}

func (db *DBLogger) UpdatePrivacyDB(group *Group) {
	db.Model(group).Select("is_private").Update("IsPrivate", group.IsPrivate)
	db.Model(group).Select("privacy_room_id").Update("PrivacyRoomID", group.PrivacyRoomID)
}

func (db *DBLogger) SaveMemberAddition(group *Group) {
	db.Model(group).Update(group)
}

func (db *DBLogger) SaveMemberRemoval(group *Group, members []Member) {
	for _, member := range members {
		db.Model(group).Delete(member)
	}
}

func (db *DBLogger) GetGroupsFromDB(groupList GroupList) {
	var foundGroups []*Group
	db.Find(&foundGroups)

	for _, group := range foundGroups {
		var members []Member
		db.Model(&group).Related(&members)

		group.Members = members

		saveName := strings.ToLower(group.Name)
		groupList[saveName] = group
	}
}

func (db *DBLogger) SyncAllGroups(groups GroupList) {
	//TODO: Create workergroup to have about 10 groups to be updated at a time.
	var wg sync.WaitGroup

	for _, group := range groups {
		wg.Add(1)
		go func(group *Group, wg *sync.WaitGroup) {
			Logger.SyncGroup(group)
			wg.Done()
		}(group, &wg)
	}

	wg.Wait()
}

func (db *DBLogger) SyncGroup(group *Group) {
	var members []Member
	db.Model(&group).Related(&members)

	group.Members = members
}

func (db *DBLogger) CreateLogEntry(msgObj messageResponse) {
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