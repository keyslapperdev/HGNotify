package main

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type (
	GroupList map[string]*namedGroup
)

type namedGroup struct {
	//    ID      uint     `json:"id"`
	Name    string `json:"groupName"`
	Members []struct {
		Name string `json:"memberName"`
		GID  string `json:"gchatID"`
	} `json:"Members"`
	Origin string `json:"homeRoom"`
}

func (ng *namedGroup) addMember(member User) {
	addition := struct {
		Name string `json:"memberName"`
		GID  string `json:"gchatID"`
	}{
		Name: member.Name,
		GID:  member.GID,
	}

	ng.Members = append(ng.Members, addition)
}

func (ng *namedGroup) removeMember(member User) {
	for i, groupMember := range ng.Members {
		if member.GID == groupMember.GID {
			ng.Members = append(ng.Members[:i], ng.Members[i+1:]...)
		}
	}
}

func (gl GroupList) Create(groupName string, msgObj messageResponse) string {
	saveName, exists := gl.CheckGroup(groupName)
	if exists {
		return fmt.Sprintf("Group %q seems to already exist. If you'd like to remove and recreate the group please say \"@HGNotify delete %s\" followed by \"@HGNotify create %s @Members...\"", groupName, groupName, groupName)
	}

	var (
		mentions   = msgObj.Message.Mentions
		newGroup   = new(namedGroup)
		newMembers string

		seen = checkSeen()
	)

	newGroup.Name = groupName

	for i, mention := range mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			if i > 1 {
				newMembers += ","
			}
			newGroup.addMember(mention.Called.User)

			newMembers += " " + mention.Called.Name
		}
	}

	gl[saveName] = newGroup
	return fmt.Sprintf("Created group %q with user(s) %s", groupName, newMembers)
}

func (gl GroupList) Delete(groupName string) string {
	saveName, exists := gl.CheckGroup(groupName)
	if !exists {
		return fmt.Sprintf("Group %q doesn't seem to exist to be deleted.", groupName)
	}

	delete(gl, saveName)
	return fmt.Sprintf("Group %q has been deleted, along with all it's data.", groupName)
}

func (gl GroupList) AddMembers(groupName string, msgObj messageResponse) string {
	saveName, exists := gl.CheckGroup(groupName)
	if !exists {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	var (
		addedMembers    string
		existingMembers string
		text            string

		seen = checkSeen()
	)

	for _, mention := range msgObj.Message.Mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.CheckMember(groupName, mention.Called.GID)

			if !exist {
				gl[saveName].addMember(mention.Called.User)

				addedMembers += mention.Called.Name + " "
			} else {
				existingMembers += mention.Called.Name + " "
			}
		}
	}

	if addedMembers != "" {
		text += fmt.Sprintf("Got [ %s] added to the group %q. ", addedMembers, groupName)
	}

	if existingMembers != "" {
		text += fmt.Sprintf("\nUser(s) [ %s] already added the group %q. ", existingMembers, groupName)
	}

	return text
}

func (gl GroupList) RemoveMembers(groupName string, msgObj messageResponse) string {
	saveName, exists := gl.CheckGroup(groupName)
	if !exists {
		return fmt.Sprintf("The group %q does not seem to exist.", groupName)
	}

	var (
		removedMembers     string
		nonExistantMembers string
		text               string

		seen = checkSeen()
	)

	for _, mention := range msgObj.Message.Mentions {
		if seen(mention.Called.Name) {
			continue
		}

		if mention.Called.Type != "BOT" && mention.Type == "USER_MENTION" {
			exist := gl.CheckMember(groupName, mention.Called.GID)

			if exist {
				gl[saveName].removeMember(mention.Called.User)

				removedMembers += mention.Called.Name + " "
			} else {
				nonExistantMembers += mention.Called.Name + " "
			}
		}
	}

	if removedMembers != "" {
		text += fmt.Sprintf("I've removed [ %s] from %q. ", removedMembers, groupName)
	}

	if nonExistantMembers != "" {
		text += fmt.Sprintf("\nUser(s) [ %s] didn't seem to exist when attempting to remove them from %q. ", nonExistantMembers, groupName)
	}

	return text
}

func (gl GroupList) List(groupName string) string {
	if groupName == "" {
		return spew.Sprintf("Structure of all stored groups: ```%v```", gl)
	}

	return spew.Sprintf("Structure of requested group %q: ```%v```", groupName, gl[strings.ToLower(groupName)])
}

func (gl GroupList) CheckGroup(groupName string) (saveName string, here bool) {
	//TODO: Check for proper formatting
	here = true
	saveName = strings.ToLower(groupName)

	if _, exist := gl[saveName]; !exist {
		here = false
	}

	return
}

func (gl GroupList) CheckMember(groupName, memberID string) (here bool) {
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
