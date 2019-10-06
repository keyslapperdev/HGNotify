package main

import (
	"fmt"
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
		return fmt.Sprintf("Group %q seems to already exist.\nIf you'd like to remove and recreate the group please say \"%s disband %s\" followed by \"%s create %s @Members...\"", groupName, BOTNAME, groupName, BOTNAME, groupName)
	}

	var (
		mentions   = msgObj.Message.Mentions
		newGroup   = new(Group)
		newMembers string

		additions   int
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
			additions++
			if additions > 1 {
				newMembers += ","
			}
			newGroup.addMember(mention.Called.User)

			newMembers += " " + mention.Called.Name
			lastNameLen = len(mention.Called.Name)
		}
	}

	switch additions {
	case 0:
		newMembers = "with no users"
	case 1:
		newMembers = "with the user " + newMembers
	case 2:
		newMembers = "with users " + newMembers
		foreMemberBytes := []byte(newMembers)[:len(newMembers)-lastNameLen-2]
		afterMemberBytes := []byte(newMembers)[len(newMembers)-lastNameLen:]

		newMembers = string(foreMemberBytes) + " and " + string(afterMemberBytes)
	default:
		newMembers = "with users " + newMembers
		foreMemberBytes := []byte(newMembers)[:len(newMembers)-lastNameLen]
		afterMemberBytes := []byte(newMembers)[len(newMembers)-lastNameLen:]

		newMembers = string(foreMemberBytes) + "and " + string(afterMemberBytes)
	}

	go Logger.SaveCreatedGroup(newGroup)
	gl[saveName] = newGroup
	return fmt.Sprintf("Created group %q %s.", groupName, newMembers)
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

		added       int
		existing    int
		lastNameLen int

		seen = checkSeen()
	)

	for _, mention := range msgObj.Message.Mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.checkMember(groupName, mention.Called.GID)

			if !exist {
				added++
				gl[saveName].addMember(mention.Called.User)

				addedMembers += mention.Called.Name + " "
			} else {
				existing++
				existingMembers += mention.Called.Name + " "
			}
		}
	}

	if added == 0 && existing == 0 {
		return "No users to add. Please @ the member you'd like to add to the group."
	}

	if added > 0 {
		switch added {
		case 1:
			addedMembers = "the user " + addedMembers
		case 2:
			addedMembers = "the users " + addedMembers
			foreMemberBytes := []byte(addedMembers)[:len(addedMembers)-lastNameLen-2]
			afterMemberBytes := []byte(addedMembers)[len(addedMembers)-lastNameLen:]

			addedMembers = string(foreMemberBytes) + " and " + string(afterMemberBytes)
		default:
			addedMembers = "the users " + addedMembers
			foreMemberBytes := []byte(addedMembers)[:len(addedMembers)-lastNameLen]
			afterMemberBytes := []byte(addedMembers)[len(addedMembers)-lastNameLen:]

			addedMembers = string(foreMemberBytes) + "and " + string(afterMemberBytes)
		}

		go Logger.SaveMemberAddition(gl[saveName])
		text += fmt.Sprintf("I've added %s to the group %q.", addedMembers, groupName)
	}

	if existing > 0 {
		switch existing {
		case 1:
			existingMembers = "The user " + existingMembers
		case 2:
			existingMembers = "The users " + existingMembers
			foreMemberBytes := []byte(existingMembers)[:len(existingMembers)-lastNameLen-2]
			afterMemberBytes := []byte(existingMembers)[len(existingMembers)-lastNameLen:]

			existingMembers = string(foreMemberBytes) + " and " + string(afterMemberBytes)
		default:
			existingMembers = "The users " + existingMembers
			foreMemberBytes := []byte(existingMembers)[:len(existingMembers)-lastNameLen]
			afterMemberBytes := []byte(existingMembers)[len(existingMembers)-lastNameLen:]

			existingMembers = string(foreMemberBytes) + "and " + string(afterMemberBytes)
		}

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
				membersToRemoveDB = append(
					membersToRemoveDB,
					gl[saveName].removeMember(mention.Called.User),
				)

				removedMembers += mention.Called.Name + " "
			} else {
				nonExistantMembers += mention.Called.Name + " "
			}
		}
	}

	if removedMembers == "" && nonExistantMembers == "" {
		return "No members to remove. Please @ the member you are wanting to remove."
	}

	if removedMembers != "" {
		go Logger.SaveMemberRemoval(gl[saveName], membersToRemoveDB)
		text += fmt.Sprintf("I've removed [ %s] from %q. ", removedMembers, groupName)
	}

	if nonExistantMembers != "" {
		text += fmt.Sprintf("\nUser(s) [ %s] didn't seem to exist when attempting to remove them from %q. ", nonExistantMembers, groupName)
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

	botLen := len(BOTNAME)
	botIndex := strings.Index(message, BOTNAME)

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

		return fmt.Sprintf("Here are all of the usable group names: ```%s``` If the group is private, it will not appear in this list. Ask me about a specfic group for more information. ( %s list groupName )", string([]byte(allGroupNames)[3:]), BOTNAME)
	}

	saveName, meta := gl.checkGroup(groupName, msgObj)
	if meta != "" {
		if !strings.Contains(meta, "exist") {
			return fmt.Sprintf("Group %q does not seem to exist.", groupName)
		}

		if strings.Contains(meta, "private") {
			return fmt.Sprintf("The group %q is private, and you may not view it.", groupName)
		}
	}

	yamlList, err := yaml.Marshal(gl[saveName])
	checkError(err)

	return fmt.Sprintf("Here are details for %q: ```%s```", groupName, string(yamlList))
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

	if group.IsPrivate {
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
