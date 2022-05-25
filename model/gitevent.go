package model

import (
	"time"

	"github.com/google/uuid"
)

type Githubevent struct {
	Uuid              uuid.UUID
	CommitedBy        string
	CommitedAt        time.Time
	Repository        string
	Commitstat        string
	Availablebranches string
	Commitmessage     string
}
