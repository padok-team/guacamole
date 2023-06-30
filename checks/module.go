package checks

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

func main() {
	type Config struct {
		Foo string `hcl:"foo"`
		Baz string `hcl:"baz"`
	}

	const exampleConfig = `
	{
		"foo": "bar",
		"baz": "boop"
	}
	`

	var config Config
	err := hclsimple.Decode(
		"example.json", []byte(exampleConfig),
		nil, &config,
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	fmt.Printf("Configuration is %v\n", config)

}
