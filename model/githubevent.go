package model

type Githubevent struct {
	CommitedBy        string
	CommitedAt        string
	Repository        string
	Added             []int
	Deleted           []int
	Filename          []string
	Availablebranches []string
}
