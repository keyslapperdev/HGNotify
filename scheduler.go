package main

import "time"

// Schedule contains information required to schedule a
// message
type Schedule struct {
	ID  	     uint       `gorm:"primary_key;not null;unique"`
	SessKey      string     `gorm:"not null"`
	Recurring    bool	    `gorm:"not null"`
	DayKey       string	   
	ExecutedOn   time.Time  `gorm:"not null"`
	UpdatedOn    time.Time  `gorm:"not null"`
	CompletedOn  time.Time  `gorm:"not null"`
	GroupID      uint       `gorm:"not null"`
	MessageLabel string     `gorm:"not null"`
	MessageText  string     `gorm:"not null"`
}