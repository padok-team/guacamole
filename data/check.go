package data

type Check struct {
	ID                string
	Name              string
	Status            string
	RelatedGuidelines string
	Errors            []Error
}
