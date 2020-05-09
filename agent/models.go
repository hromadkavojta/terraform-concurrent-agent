package agent

import "sync"

type Service struct {
	PlansProvided int
	wg            sync.WaitGroup
	repo          string
	url           string
}

func NewService(
	variables ServiceVariables,
) *Service {
	return &Service{
		repo: variables.Repo,
		url:  variables.Url,
	}
}

type ServiceVariables struct {
	Repo string
	Url  string
}

type ApplyStruct struct {
	Plan string `json:"plan"`
}
