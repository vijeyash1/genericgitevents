package model

import "time"

type Githubevent struct {
	CommitedBy        string
	CommitedAt        time.Time
	Repository        string
	Added             []int
	Deleted           []int
	Filename          []string
	Availablebranches []string
	Commitmessage     string
}
