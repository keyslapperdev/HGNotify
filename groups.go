package main

import (
	"fmt"
	"strings"
)

type (
	GroupList map[string]namedGroup
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
	saveName := strings.ToLower(groupName)
	if _, exists := gl[saveName]; exists {
		return fmt.Sprintf("Group %q seems to already exist. If you'd like to remove and recreate the group please say \"@HGNotify delete %s\" followed by \"@HGNotify create %s @Members...\"", groupName, groupName, groupName)
	}

	mentions := msgObj.Message.Mentions
	newGroup := namedGroup{Name: groupName}

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
	saveName := strings.ToLower(groupName)
	if _, exists := gl[saveName]; !exists {
		return fmt.Sprintf("Group %q doesn't seem to exist to be deleted.", groupName)
	}

	delete(gl, groupName)
	return fmt.Sprintf("Group %q has been deleted, along with all it's data.", groupName)
}

func (gl GroupList) List(groupName string) string {
	saveName := strings.ToLower(groupName)

	if groupName == "" {
		return fmt.Sprintf("Structure of all stored groups: ```%v```", gl)
	}

	return fmt.Sprintf("Structure of requested group %q: ```%v```", groupName, gl[saveName])
}

func (gl GroupList) Validate(groupName string) (name, msg string, ok bool) {
	ok = true

	if _, exist := Groups[groupName]; !exist {
		msg = fmt.Sprintf("Group %q doesn't seem to exist yet, try initializing it with \"@HGNotify create %s\".", groupName, groupName)
		ok = false
	} else {
		name = groupName
	}

	return
}
