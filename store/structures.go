package store

import "time"

type Program struct {
	Id int64
	Name string
	CreatedAt time.Time
}