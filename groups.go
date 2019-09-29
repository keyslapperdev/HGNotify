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

func (gl GroupList) Create(groupName string, msgObj messageResponse) string {
	saveName, exists := gl.CheckGroup(groupName)
	if exists {
		return fmt.Sprintf("Group %q seems to already exist. If you'd like to remove and recreate the group please say \"@HGNotify delete %s\" followed by \"@HGNotify create %s @Members...\"", groupName, groupName, groupName)
	}

	mentions := msgObj.Message.Mentions
	newGroup := new(namedGroup)
	newGroup.Name = groupName

	var newMembers string
	for i, mention := range mentions {
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

func (gl GroupList) AddMember(groupName string, msgObj messageResponse) string {
	saveName, exists := gl.CheckGroup(groupName)
	if !exists {
		return fmt.Sprintf("Group %q does not seem to exist.", groupName)
	}

	var (
		addedMembers    string
		existingMembers string
		text            string
	)

	for _, mention := range msgObj.Message.Mentions {
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
		text += fmt.Sprintf("User(s) [ %s] previously added the group %q. ", existingMembers, groupName)
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
	here = true
	saveName := strings.ToLower(groupName)

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
