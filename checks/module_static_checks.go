package checks

import (
	"strings"
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
			// Create temporary slice because we may be deleting some elements from the original
			// Compare checks and their possible whitelisting via comments
			// for _, checkError := range check.Errors {
			for i := len(check.Errors) - 1; i >= 0; i-- {
				for _, whitelisterror := range whitelistComments {
					// We check the line number +1 because the comment is always above the code block
					if strings.Contains(check.Errors[i].Path, whitelisterror.Path) && check.Errors[i].LineNumber < whitelisterror.LineNumber+4 && check.ID == whitelisterror.CheckID {
						check.Errors = append(check.Errors[:i], check.Errors[i+1:]...)
						break
					}
				}
			}
			// Replace the check error with the array after whitelisting
			if len(check.Errors) == 0 {
				check.Status = "✅"
			}
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
