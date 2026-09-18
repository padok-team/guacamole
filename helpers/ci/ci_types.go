package ci

type scope uint8

const (
	scopeLayer scope = iota
	scopeModule
)

func (s scope) String() string {
	return [...]string{"layer", "module"}[s]
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
	projectDirAbs string
	baseBranch    string
	mrSHA         string
	postComment   bool
}

type totals struct {
	overallPass  int
	overallTotal int
	hasError     bool
}
