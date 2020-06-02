package main

import (
	"helpers"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

//Test Helper data
func genRandName(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return helpers.RandString(length)
}

func genUserGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "users/" + helpers.StringWithCharset(length, "0123456789")
}

func genRoomGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "spaces/" + helpers.StringWithCharset(length, "0123456789")
}
