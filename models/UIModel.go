package models

import "time"

type Model struct {
	Header       string
	Choices      []string         // items on the to-do list
	Cursor       int              // which to-do list item our cursor is pointing at
	Selected     map[int]struct{} // which to-do items are selected
	Footer       string
	TickDuration time.Duration
}
