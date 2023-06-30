package data

type Check struct {
	Name              string
	Status            string
	RelatedGuidelines string
	Errors            []string
}

type Module struct {
	Name     string
	FullPath string
}
