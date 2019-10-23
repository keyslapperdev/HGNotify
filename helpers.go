package main

import (
	"github.com/davecgh/go-spew/spew"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func describe(msg string, v ...interface{}) {
	spew.Printf(msg, v...)
}

func dump(v ...interface{}) {
	spew.Dump(v...)
}
