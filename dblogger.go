package main

import (
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func startDBLogger() (db *gorm.DB) {
	db, err := gorm.Open("mysql", "z_hgnotify_user:z_hgnotify_password@/z_hgnotify_test?charset=utf8mb4&parseTime=True")
	checkError(err)

	db.LogMode(true)
	return
}

func setupTables(db *gorm.DB) {
	db.AutoMigrate(&Group{}, &Member{})
	db.Model(&Member{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
}

func saveCreatedGroup(db *gorm.DB, ng *Group) {
	db.Create(ng)
}

func getGroupsFromDB(db *gorm.DB, groupList GroupList) {
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
