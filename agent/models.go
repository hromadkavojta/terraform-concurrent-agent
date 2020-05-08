package agent

type Service struct {
	planned       [][]string
	processing    [][]string
	PlansProvided int
}

type ApplyStruct struct {
	Plan string `json:"plan"`
}
