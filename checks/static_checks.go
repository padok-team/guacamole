package checks

import (
	"guacamole/data"
	"sync"
)

func StaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func() (data.Check, error){
		"ProviderInModule":        ProviderInModule,
		"Stuttering":              Stuttering,
		"SnakeCase":               SnakeCase,
		"MissingVarDescription":   MissingVarDescription,
		"CollectionVarNamePlural": VarNumberMatchesType,
		"VariableTypeAny":         VariableTypeAny,
	}

	var checkResults []data.Check

	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for _, checkFunction := range checks {
		go func(checkFunction func() (data.Check, error)) {
			defer wg.Done()

			check, err := checkFunction()
			if err != nil {
				panic(err)
			}
			c <- check
		}(checkFunction)
	}

	wg.Wait()

	for i := 0; i < len(checks); i++ {
		checkResults = append(checkResults, <-c)
	}

	return checkResults
}
