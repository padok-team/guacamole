package checks

import (
	"guacamole/data"
	"sync"

	"golang.org/x/exp/slices"
)

func StaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func() (data.Check, error){
		"ProviderInModule":       ProviderInModule,
		"Stuttering":             Stuttering,
		"SnakeCase":              SnakeCase,
		"VarContainsDescription": VarContainsDescription,
		"VarNumberMatchesType":   VarNumberMatchesType,
		"VarTypeAny":             VarTypeAny,
		"RemoteModuleVersion":    RemoteModuleVersion,
		"RequiredProviderVersionOperatorInModules": RequiredProviderVersionOperatorInModules,
		"ResourceNamingThisThese":                  ResourceNamingThisThese,
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
		check := <-c
		checkResults = append(checkResults, check)
	}

	// Sort the checks by their ID
	slices.SortFunc(checkResults, func(i, j data.Check) int {
		if i.ID < j.ID {
			return -1
		}
		if i.ID > j.ID {
			return 1
		}
		return 0
	})

	return checkResults
}
