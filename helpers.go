package main

import (
	"math/rand"
	"time"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

//Test helper data
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return StringWithCharset(length, charset)
}

//Test Helper data
func genRandName(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return RandString(length)
}

func genUserGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "users/" + StringWithCharset(length, "0123456789")
}

func genRoomGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "spaces/" + StringWithCharset(length, "0123456789")
}
