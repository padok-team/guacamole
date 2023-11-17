package checks

import (
	"sync"

	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"

	"golang.org/x/exp/slices"
)

func ModuleStaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func(m []data.TerraformModule) (data.Check, error){
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

	// Find recusively all the modules in the current directory
	modules, whitelistComments, err := helpers.GetModules()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for _, checkFunction := range checks {
		go func(checkFunction func(m []data.TerraformModule) (data.Check, error)) {
			defer wg.Done()

			check, err := checkFunction(modules)
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
