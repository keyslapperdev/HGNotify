package main

import (
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
	db, err := gorm.Open(
		conf.Driver,
		conf.DBUser+":"+
			conf.DBPass+"@/"+
			conf.DBName+"?"+
			conf.DBOpts)

	checkError(err)

	return DBLogger{db}
}

//SetupTables method is used when the file first starts up to
//ensure all databases are created and updated as they should be.
//If everything is fine, this is only ran once, ever.
func (db *DBLogger) SetupTables() {
	db.AutoMigrate(&Group{}, &Member{}, &NotifyLog{})
	db.Model(&Member{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "RESTRICT")
}

//SaveCreatedGroup method is used to update the database whenever
//a new group is created
func (db *DBLogger) SaveCreatedGroup(group *Group) {
	db.Create(group)
}

//DisbandGroup method will delete a group's entry from the database,
//along with all the associated users.
func (db *DBLogger) DisbandGroup(group *Group) {
	db.Unscoped().Delete(group)
}

//UpdatePrivacyDB method toggles the privacy settings for the specified
//group. It's a bit different because when the restriction is removed, the
//values entered into the database are "zero value", so gorm ignores them.
//To get them to set the zero value I have to be specific with the query.
func (db *DBLogger) UpdatePrivacyDB(group *Group) {
	db.Model(group).Select("is_private").Update("IsPrivate", group.IsPrivate)
	db.Model(group).Select("privacy_room_id").Update("PrivacyRoomID", group.PrivacyRoomID)
}

//SaveMemberAddition method adds a member to the associated group
func (db *DBLogger) SaveMemberAddition(group *Group) {
	db.Model(group).Update(group)
}

//SaveMemberRemoval method marks the assocaited memeber as removed from
//the group.
func (db *DBLogger) SaveMemberRemoval(group *Group, members []Member) {
	for _, member := range members {
		db.Model(group).Delete(member)
	}
}

//GetGroupsFromDB method syncs the database groups to the in-memory group list
//this is ran when the program starts up.
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

//SyncAllGroups method syncs the database groups to the in-memory group list
//during runtime. Just in case.
func (db *DBLogger) SyncAllGroups(groups GroupList) {
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
	var members []Member
	db.Model(&group).Related(&members)

	group.Members = members
}

//CreateLogEntry method logs usage of the bot to the database.
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
