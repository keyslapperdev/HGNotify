package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/jinzhu/gorm"
)

type (
	GroupList map[string]*Group
)

type Group struct {
	gorm.Model    `yaml:"-"`
	Name          string   `yaml:"groupName" gorm:"not null"`
	Members       []Member `yaml:"members" gorm:"foreignkey:GroupID"`
	IsPrivate     bool     `yaml:"private" gorm:"default:false;not null"`
	PrivacyRoomID string   `yaml:"-"`
}

type Member struct {
	gorm.Model `yaml:"-"`
	GroupID    uint   `yaml:"-" gorm:"index:idx_members_group_id"`
	Name       string `yaml:"memberName" gorm:"not null"`
	GID        string `yaml:"gchatID" gorm:"not null"`
}

func (ng *Group) addMember(member User) {
	addition := Member{
		Name: member.Name,
		GID:  member.GID,
	}

	ng.Members = append(ng.Members, addition)
}

func (ng *Group) removeMember(member User) (removed Member) {
	for i, groupMember := range ng.Members {
		if member.GID == groupMember.GID {
			removed = groupMember
			ng.Members = append(ng.Members[:i], ng.Members[i+1:]...)
		}
	}

	return
}

func (gl GroupList) Create(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "name") {
		return fmt.Sprintf("Cannot use %q as group name. Group names can contain letters, numbers, underscores, and dashes, maximum length is 40 characters", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q already exists and is private.", groupName)
	}

	if strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q seems to already exist.\nIf you'd like to remove and recreate the group please say \"%s disband %s\" followed by \"%s create %s @Members...\"", groupName, BotName, groupName, BotName, groupName)
	}

	var (
		mentions   = msgObj.Message.Mentions
		newGroup   = new(Group)
		newMembers string

		numAdded    int
		lastNameLen int

		seen = checkSeen()
	)

	newGroup.Name = groupName
	newGroup.IsPrivate = false

	for _, mention := range mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			numAdded++
			if numAdded > 1 {
				newMembers += ","
			}
			newGroup.addMember(mention.Called.User)

			newMembers += " " + mention.Called.Name
			lastNameLen = len(mention.Called.Name)
		}
	}

	if numAdded == 0 {
		newMembers = "no users"
	} else {
		newMembers = grammar(newMembers, numAdded, lastNameLen)
	}

	go Logger.SaveCreatedGroup(newGroup)
	gl[saveName] = newGroup
	return fmt.Sprintf("Created group %q with %s.", groupName, newMembers)
}

func (gl GroupList) Disband(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not mutate it.", groupName)
	}

	go Logger.DisbandGroup(gl[saveName])
	delete(gl, saveName)
	return fmt.Sprintf("Group %q has been deleted, along with all it's data.", groupName)
}

func (gl GroupList) AddMembers(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not mutate it.", groupName)
	}

	var (
		addedMembers    string
		existingMembers string
		text            string

		numAdded         int
		numExist         int
		lastAddedNameLen int
		lastExistNameLen int

		seen = checkSeen()
	)

	for _, mention := range msgObj.Message.Mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.checkMember(groupName, mention.Called.GID)

			if !exist {
				numAdded++
				if numAdded > 1 {
					addedMembers += ","
				}
				gl[saveName].addMember(mention.Called.User)

				addedMembers += " " + mention.Called.Name
				lastAddedNameLen = len(mention.Called.Name)
			} else {
				numExist++
				if numExist > 1 {
					existingMembers += ","
				}

				existingMembers += " " + mention.Called.Name
				lastExistNameLen = len(mention.Called.Name)
			}
		}
	}

	if numAdded == 0 && numExist == 0 {
		return "No users to add. Please @ the member you'd like to add to the group."
	}

	if numAdded > 0 {
		addedMembers = grammar(addedMembers, numAdded, lastAddedNameLen)

		go Logger.SaveMemberAddition(gl[saveName])
		text += fmt.Sprintf("I've added %s to the group %q.", addedMembers, groupName)
	}

	if numExist > 0 {
		existingMembers = grammar(addedMembers, numExist, lastExistNameLen)

		text += fmt.Sprintf("\n%s already added the group %q. ", existingMembers, groupName)
	}

	return text
}

func (gl GroupList) RemoveMembers(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not mutate it.", groupName)
	}

	var (
		removedMembers     string
		nonExistantMembers string

		membersToRemoveDB []Member

		numNonExist         int
		numRemoved          int
		lastRemovedNameLen  int
		lastNonExistNameLen int

		text string

		seen = checkSeen()
	)

	for _, mention := range msgObj.Message.Mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.checkMember(groupName, mention.Called.GID)

			if exist {
				numRemoved++
				if numRemoved > 1 {
					removedMembers += ","
				}

				membersToRemoveDB = append(
					membersToRemoveDB,
					gl[saveName].removeMember(mention.Called.User),
				)

				removedMembers += " " + mention.Called.Name
				lastRemovedNameLen = len(mention.Called.Name)
			} else {
				numNonExist++
				if numNonExist > 1 {
					nonExistantMembers += ","
				}

				nonExistantMembers += " " + mention.Called.Name
				lastNonExistNameLen = len(mention.Called.Name)
			}
		}
	}

	if numRemoved == 0 && numNonExist == 0 {
		return "No members to remove. Please @ the member you are wanting to remove."
	}

	if numRemoved > 0 {
		removedMembers = grammar(removedMembers, numRemoved, lastRemovedNameLen)

		go Logger.SaveMemberRemoval(gl[saveName], membersToRemoveDB)
		text += fmt.Sprintf("I've removed %s from %q. ", removedMembers, groupName)
	}

	if numNonExist > 0 {
		nonExistantMembers = grammar(nonExistantMembers, numNonExist, lastNonExistNameLen)

		text += fmt.Sprintf("\n%s didn't seem to exist when attempting to remove them from %q. ", nonExistantMembers, groupName)
	}

	return text
}

func (gl GroupList) Restrict(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not mutate it.", groupName)
	}

	if gl[saveName].IsPrivate {
		gl[saveName].IsPrivate = false
		gl[saveName].PrivacyRoomID = ""

		go Logger.UpdatePrivacyDB(gl[saveName])
		return fmt.Sprintf("I've set %q to public, now it can be used in any room.", groupName)
	}

	gl[saveName].IsPrivate = true
	gl[saveName].PrivacyRoomID = msgObj.Room.GID

	go Logger.UpdatePrivacyDB(gl[saveName])
	return fmt.Sprintf("I've set %q to be private, the group can only be used in this room now.", groupName)
}

func (gl GroupList) Notify(groupName string, msgObj messageResponse) string {
	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not use it.", groupName)
	}

	var memberList string
	//TODO: Check if users are in the room before adding them to list
	for _, member := range gl[saveName].Members {
		memberList += "<" + member.GID + "> "
	}

	message := msgObj.Message.Text

	botLen := len(BotName)
	botIndex := strings.Index(message, BotName)

	tmpMessage := string([]byte(message)[botLen+botIndex:])

	groupLen := len(groupName)
	groupIndex := strings.Index(tmpMessage, groupName)

	newMessage := fmt.Sprintf("%s said:\n\n%s",
		msgObj.Message.Sender.Name,
		strings.Replace(
			message,
			string([]byte(message)[botIndex:botIndex+botLen+groupIndex+groupLen]),
			memberList,
			1,
		),
	)

	if len(newMessage) >= 4000 {
		return "My apologies, your message with the group added would exceed Google Chat's character limit. :("
	}

	return newMessage
}

func (gl GroupList) List(groupName string, msgObj messageResponse) string {
	if groupName == "" {
		noneToShow := "There are no groups to show currently. :("

		if len(gl) == 0 {
			return noneToShow
		}

		var allGroupNames string
		for name := range gl {
			_, meta := gl.checkGroup(name, msgObj)
			if !strings.Contains(meta, "private") {
				allGroupNames += " | " + gl[name].Name
			}
		}

		if len(allGroupNames) == 0 {
			return noneToShow
		}

		return fmt.Sprintf("Here are all of the usable group names: ```%s``` If the group is private, it will not appear in this list. Ask me about a specfic group for more information. ( %s list groupName )", string([]byte(allGroupNames)[3:]), BotName)
	}

	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	if strings.Contains(meta, "private") {
		return fmt.Sprintf("The group %q is private, and you may not view it.", groupName)
	}

	yamlList, err := yaml.Marshal(gl[saveName])
	checkError(err)

	return fmt.Sprintf("Here are details for %q: ```%s```", groupName, string(yamlList))
}

func (gl GroupList) SyncGroupMembers(groupName string, msgObj messageResponse) string {
	if !msgObj.IsMaster {
		return "Invalid option received. I'm not sure what to do about \"syncgroup\"."
	}

	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	oldMembers := gl[saveName].Members
	Logger.SyncGroup(gl[saveName])
	syncedMembers := gl[saveName].Members

	text := fmt.Sprintf("Group %q synced, ", groupName)

	if reflect.DeepEqual(oldMembers, syncedMembers) {
		text += "and no changes were made."
	} else {
		text += "and some changes were made."
	}

	return text
}

func (gl GroupList) SyncAllGroups(msgObj messageResponse) string {
	if !msgObj.IsMaster {
		return "Invalid option received. I'm not sure what to do about \"syncallgroups\"."
	}

	Logger.SyncAllGroups(gl)

	fmt.Println("All groups synced")
	return "All groups synced."
}

func (gl GroupList) checkGroup(groupName string, msgObj messageResponse) (saveName, meta string) {
	match, err := regexp.Match(`^[\w-]{0,40}$`, []byte(groupName))
	checkError(err)

	if match {
		meta += "name"
	} else {
		return
	}

	saveName = strings.ToLower(groupName)
	group, exist := gl[saveName]

	if exist {
		meta += "exist"
	} else {
		return
	}

	if group.IsPrivate && !msgObj.IsMaster {
		if group.PrivacyRoomID != msgObj.Room.GID {
			meta += "private"
		}
	}

	return
}

func (gl GroupList) checkMember(groupName, memberID string) (here bool) {
	saveName := strings.ToLower(groupName)

	if len(gl[saveName].Members) == 0 {
		here = false
	} else {
		here = true
	}

	for _, member := range gl[saveName].Members {
		if memberID == member.GID {
			here = true
			break
		} else {
			here = false
		}
	}

	return
}

func grammar(members string, delta, lastNameLen int) (corrected string) {
	switch delta {
	case 1:
		corrected = "the user" + members
	case 2:
		corrected = "the users" + members
		foreMemberBytes := []byte(corrected)[:len(corrected)-lastNameLen-2]
		afterMemberBytes := []byte(corrected)[len(corrected)-lastNameLen:]

		corrected = string(foreMemberBytes) + " and " + string(afterMemberBytes)
	default:
		corrected = "the users" + members
		foreMemberBytes := []byte(corrected)[:len(corrected)-lastNameLen]
		afterMemberBytes := []byte(corrected)[len(corrected)-lastNameLen:]

		corrected = string(foreMemberBytes) + "and " + string(afterMemberBytes)
	}

	return
}

func checkSeen() func(name string) bool {
	var seenMembers []string

	return func(name string) bool {
		for _, seenMember := range seenMembers {
			if seenMember == name {
				return true
			}
		}

		seenMembers = append(seenMembers, name)
		return false
	}
}

func isGroup(groupName string) bool {
	_, exists := Groups[strings.ToLower(groupName)]

	if exists {
		return true
	}

	return false
}
