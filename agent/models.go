package agent

import "sync"
import git "github.com/go-git/go-git/v5"

type Service struct {
	PlansProvided int
	wg            sync.WaitGroup
	repo          string
	url           string
	//r             *git.Repository
	AccessToken    string
	w              *git.Worktree
	committer      string
	committerEmail string
}

func NewService(
	variables ServiceVariables,
) *Service {
	return &Service{
		repo: variables.Repo,
		url:  variables.Url,
		//r:             variables.Repository,
		AccessToken:    variables.AccessToken,
		committer:      variables.Committer,
		committerEmail: variables.CommitterEmail,
	}
}

type ServiceVariables struct {
	Repo           string
	Url            string
	Repository     *git.Repository
	AccessToken    string
	Committer      string
	CommitterEmail string
}

type ApplyStruct struct {
	Plan string `json:"plan"`
}
