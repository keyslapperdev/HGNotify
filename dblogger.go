package main

import (
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DBLogger struct {
	*gorm.DB
}

func startDBLogger() DBLogger {
	db, err := gorm.Open("mysql", "z_hgnotify_user:z_hgnotify_password@/z_hgnotify_test?charset=utf8mb4&parseTime=True")
	checkError(err)

	db.LogMode(true)
	return DBLogger{db}
}

func (db *DBLogger) SetupTables() {
	db.AutoMigrate(&Group{}, &Member{})
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
