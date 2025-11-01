package models

import "time"

type Model struct {
	Header         string
	Choices        []string // items on the to-do list
	Cursor         int      // which to-do list item our cursor is pointing at
	Selected       int      // which to-do items are selected
	AdditionalInfo string   // additional information about the selected item
	Footer         string
	TickDuration   time.Duration
}
