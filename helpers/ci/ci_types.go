package ci

type scope uint8

const (
	scopeLayer scope = iota
	scopeModule
)

func (s scope) String() string {
	return [...]string{"layer", "module"}[s]
}

type platform uint8

const (
	platformGitlab platform = iota
	platformGithub
)

func (p platform) String() string {
	return [...]string{"gitlab", "github"}[p]
}

type scopeDirs struct {
	scope scope
	dirs  []string
}

type result struct {
	scope       scope
	path        string
	passed      int
	total       int
	score       string
	failingText string
	hasError    bool
}

type config struct {
	platform      platform
	projectDirAbs string
	baseBranch    string
	mrSHA         string
	postComment   bool
	scanAll       bool
	failOnError   bool
}

type totals struct {
	overallPass  int
	overallTotal int
	hasError     bool
}
