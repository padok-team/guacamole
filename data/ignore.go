package data

type IgnoreComment struct {
	CheckID    string
	LineNumber int
	Path       string
}

type IgnoreModule struct {
	CheckID    string
	ModulePath string
}
