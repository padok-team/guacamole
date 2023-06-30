package data

type Check struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Condition   string `yaml:"condition"`
}
