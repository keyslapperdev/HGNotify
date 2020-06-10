package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/jinzhu/gorm"
)

//GroupList tpye is used to hold all group information in memory for
//speedy interaction with the groups.
type (
	GroupList map[string]*Group
)

//Group struct used to hold/model a group's structure. Note: Elements
//with the `yaml:"-"` tag do not appear in the 'List' function
type Group struct {
	gorm.Model    `yaml:"-"`
	Name          string   `yaml:"groupName" gorm:"not null"`
	Members       []Member `yaml:"members" gorm:"foreignkey:GroupID"`
	IsPrivate     bool     `yaml:"private" gorm:"default:false;not null"`
	PrivacyRoomID string   `yaml:"-"`
}

//Member struct used to define member information
type Member struct {
	gorm.Model `yaml:"-"`
	GroupID    uint   `yaml:"-" gorm:"index:idx_members_group_id"`
	Name       string `yaml:"memberName" gorm:"not null"`
	GID        string `yaml:"gchatID" gorm:"not null"`
}

func (g *Group) manageMember(action string, memberList *string, delta, lastNameLen *int, user User) (memberToRemoveDB Member) {
	*delta++
	if *delta > 1 {
		*memberList += ","
	}

	switch action {
	case "add":
		g.addMember(user)
	case "remove":
		memberToRemoveDB = g.removeMember(user)
	default:
		//do nothing
	}

	*memberList += " " + user.Name
	*lastNameLen = len(user.Name)
	return
}

//addMember is an unexported method, because nothing outside of this file
//uses these methods. If the name of the method wasn't clear enough, it's
//used specifcially to map a user to a member and add them to the associated
//group
func (g *Group) addMember(member User) {
	addition := Member{
		Name: member.Name,
		GID:  member.GID,
	}

	g.Members = append(g.Members, addition)
}

//removeMember is also unexported, for the same reason. This method removes
//users from the assocaited group I kind of had to get creative with this one.
//The return value is specifically so the removal can be reflected in the
//database
func (g *Group) removeMember(member User) (removed Member) {
	for i, groupMember := range g.Members {
		if member.GID == groupMember.GID {
			removed = groupMember
			g.Members = append(g.Members[:i], g.Members[i+1:]...)
		}
	}

	return
}

//Create method initializes a single group.
func (gl GroupList) Create(groupName, self string, msgObj messageResponse) string {
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
		user := mention.Called.User

		if seen(user.Name) {
			continue
		}

		if user.Type != "BOT" && mention.Type == "USER_MENTION" {
			newGroup.manageMember("add", &newMembers, &numAdded, &lastNameLen, user)
		}
	}

	if self != "" {
		newGroup.manageMember("add", &newMembers, &numAdded, &lastNameLen, msgObj.Message.Sender.User)
	}

	if numAdded == 0 {
		newMembers = "no users"
	} else {
		newMembers = correctGP(newMembers, numAdded, lastNameLen)
	}

	go Logger.SaveCreatedGroup(newGroup)
	gl[saveName] = newGroup
	return fmt.Sprintf("Created group %q with %s.", groupName, newMembers)
}

//Disband method will remove a group from the list, as well, delete the group from the
//database. The removal from the database will also remove the associated member entries
//something to be aware of.
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
	return fmt.Sprintf("Group %q has been deleted, along with all its data.", groupName)
}

//AddMembers method adds a list of members to the specified group.
func (gl GroupList) AddMembers(groupName, self string, msgObj messageResponse) string {
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
		user := mention.Called.User

		if seen(user.Name) {
			continue
		}

		if user.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.checkMember(groupName, user.GID)

			if !exist {
				gl[saveName].manageMember("add", &addedMembers, &numAdded, &lastAddedNameLen, user)
			} else {
				gl[saveName].manageMember("none", &existingMembers, &numExist, &lastExistNameLen, user)
			}
		}
	}

	if self != "" {
		sender := msgObj.Message.Sender
		exist := gl.checkMember(groupName, sender.GID)

		if !exist {
			gl[saveName].manageMember("add", &addedMembers, &numAdded, &lastAddedNameLen, sender.User)
		} else {
			gl[saveName].manageMember("none", &existingMembers, &numExist, &lastExistNameLen, sender.User)
		}
	}

	if numAdded == 0 && numExist == 0 {
		return "No users to add. Please @ the member you'd like to add to the group."
	}

	if numAdded > 0 {
		addedMembers = correctGP(addedMembers, numAdded, lastAddedNameLen)

		go Logger.SaveMemberAddition(gl[saveName])
		text += fmt.Sprintf("I've added the %s to the group %q.", addedMembers, groupName)
	}

	if numExist > 0 {
		existingMembers = correctGP(existingMembers, numExist, lastExistNameLen)

		text += fmt.Sprintf("\nThe %s already added the group %q. ", existingMembers, groupName)
	}

	return text
}

//RemoveMembers method removes the member from the group. When the member is removed from
//the group, they are not deleted from the database, but are marked as removed, just in
//case
func (gl GroupList) RemoveMembers(groupName, self string, msgObj messageResponse) string {
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
		user := mention.Called.User

		if seen(mention.Called.Name) {
			continue
		}

		if user.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.checkMember(groupName, user.GID)

			if exist {
				membersToRemoveDB = append(
					membersToRemoveDB,
					gl[saveName].manageMember(
						"remove",
						&removedMembers,
						&numRemoved,
						&lastRemovedNameLen,
						user,
					),
				)
			} else {
				gl[saveName].manageMember(
					"none",
					&nonExistantMembers,
					&numNonExist,
					&lastNonExistNameLen,
					user,
				)
			}
		}
	}

	if self != "" {
		sender := msgObj.Message.Sender
		exist := gl.checkMember(groupName, sender.GID)

		if exist {
			membersToRemoveDB = append(
				membersToRemoveDB,
				gl[saveName].manageMember(
					"remove",
					&removedMembers,
					&numRemoved,
					&lastRemovedNameLen,
					sender.User,
				),
			)
		} else {
			gl[saveName].manageMember(
				"none",
				&nonExistantMembers,
				&numNonExist,
				&lastNonExistNameLen,
				sender.User,
			)
		}
	}

	if numRemoved == 0 && numNonExist == 0 {
		return "No members to remove. Please @ the member you are wanting to remove."
	}

	if numRemoved > 0 {
		removedMembers = correctGP(removedMembers, numRemoved, lastRemovedNameLen)

		go Logger.SaveMemberRemoval(gl[saveName], membersToRemoveDB)
		text += fmt.Sprintf("I've removed the %s from %q. ", removedMembers, groupName)
	}

	if numNonExist > 0 {
		nonExistantMembers = correctGP(nonExistantMembers, numNonExist, lastNonExistNameLen)

		text += fmt.Sprintf("\nThe %s didn't seem to exist when attempting to remove them from %q. ", nonExistantMembers, groupName)
	}

	return text
}

//Restrict method restricts the interaction of the group to the room this was called in.
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

//Notify method is the bread and butter of this bot. It's it will take your message, and
//replace the botname and specified group, with the users in the list.
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
			//This bit of nonsense is how I actually do the replacing. I cast the message
			//string to an array of bytes, go to the beginning of the bot, select up until
			//the end of the group name, then cast that bit into a string to be replaced
			//by the memberList. It's gross, but efficient. Like me?
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

//List method will show you either a list of all of the groups available for use, or details
//about a specific group, depending on the options with which you call the method.
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

		//the [3:] in the byte slice cast is just to remove the leading pipe and accompanying
		//spaces
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

//SyncGroupMembers is a hidden route for the bot's admin. It's purpose is to sync the
//in-memory groups with the information in the database. Typically the database is only used
//for logging, and not really influencing the in-memory group list. However, there are some
//special circumstances which may call for manual intervention. With this method, the admin
//would be able to modify the database manually, then call this method to sync a single group.
func (gl GroupList) SyncGroupMembers(groupName string, msgObj messageResponse) string {
	if !msgObj.IsMaster {
		return "Invalid option received. I'm not sure what to do about \"syncgroup\"."
	}

	saveName, meta := gl.checkGroup(groupName, msgObj)
	if !strings.Contains(meta, "exist") {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	//Storing old member list to check later if a change has occured
	oldMembers := gl[saveName].Members
	Logger.SyncGroup(gl[saveName])
	syncedMembers := gl[saveName].Members

	text := fmt.Sprintf("Group %q synced, ", groupName)

	//At this point, I don't believe I am too concerned with the specific
	//members so much as, if a change occured. Maybe in the future, this
	//can be updated to be more specific.
	if reflect.DeepEqual(oldMembers, syncedMembers) {
		text += "and no changes were made."
	} else {
		text += "and some changes were made."
	}

	return text
}

//SyncAllGroups is similar to the philosophy of the above method. The main difference, is this
//one does a sync for all of the groups, as opposed to just one.
func (gl GroupList) SyncAllGroups(msgObj messageResponse) string {
	if !msgObj.IsMaster {
		return "Invalid option received. I'm not sure what to do about \"syncallgroups\"."
	}

	Logger.SyncAllGroups(gl)

	fmt.Println("All groups synced")
	return "All groups synced."
}

//checkGroups method checks the group and returns data about the group to be processed and
//responded to accordingly
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

	//Nothing is private for bot admin.
	if group.IsPrivate && !msgObj.IsMaster {
		if group.PrivacyRoomID != msgObj.Room.GID {
			meta += "private"
		}
	}

	return
}

//checkMember method pretty much checks solely to see if the member exists. As I'm writing
//I'm realizing this should actually be a method of the group, and not the group list. :|
//I'll do that later.
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

//correctGP is a function that appropriately adds commas and the word "and" where needed
//It's kinda gross, but it works :D
func correctGP(members string, delta, lastNameLen int) (corrected string) {
	switch delta {
	case 1:
		corrected = "user" + members
	case 2:
		corrected = "users" + members
		//The -2 at the end of lastNameLen is to remove a comma placed between
		//the two user names, as it's not needed
		foreMemberBytes := []byte(corrected)[:len(corrected)-lastNameLen-2]
		afterMemberBytes := []byte(corrected)[len(corrected)-lastNameLen:]

		corrected = string(foreMemberBytes) + " and " + string(afterMemberBytes)
	default:
		corrected = "users" + members
		foreMemberBytes := []byte(corrected)[:len(corrected)-lastNameLen]
		afterMemberBytes := []byte(corrected)[len(corrected)-lastNameLen:]

		corrected = string(foreMemberBytes) + "and " + string(afterMemberBytes)
	}

	return
}

//checkSeen is a function that checks to see if the same name was placed more than once
//when listing users for any of the methods
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

//isGroup is a function created for the Notify method. This is what checks the string after
//the bot's name call to check if it's a group. In the future, this will be called more than
//once depending on if the preceeding string was a group. This would be support for notifying
//multiple groups at once. That does seem like something useful, but not really at this time.
func isGroup(groupName string) bool {
	_, exists := Groups[strings.ToLower(groupName)]

	if exists {
		return true
	}

	return false
}
