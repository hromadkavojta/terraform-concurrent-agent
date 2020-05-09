package agent

import (
	"gopkg.in/src-d/go-git.v4"
	"sync"
)

type Service struct {
	PlansProvided int
	wg            sync.WaitGroup
	Repo          *git.Repository
}

type ApplyStruct struct {
	Plan string `json:"plan"`
}
