package data

type WhitelistComment struct {
	CheckID    string
	LineNumber int
	Path       string
}

type WhitelistModule struct {
	CheckID    string
	ModulePath string
}
