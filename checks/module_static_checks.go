package checks

import (
	"sync"

	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"

	"golang.org/x/exp/slices"
)

func ModuleStaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func(m map[string]data.TerraformModule) (data.Check, error){
		"ProviderInModule":       ProviderInModule,
		"Stuttering":             Stuttering,
		"SnakeCase":              SnakeCase,
		"VarContainsDescription": VarContainsDescription,
		"VarNumberMatchesType":   VarNumberMatchesType,
		"VarTypeAny":             VarTypeAny,
		"RemoteModuleVersion":    RemoteModuleVersion,
		"RequiredProviderVersionOperatorInModules": RequiredProviderVersionOperatorInModules,
		"ResourceNamingThisThese":                  ResourceNamingThisThese,
		"ResourceNaming":                           ResourceNaming,
	}

	var checkResults []data.Check

	// Find recusively all the modules in the current directory
	modules, err := helpers.GetModules()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for _, checkFunction := range checks {
		go func(checkFunction func(m map[string]data.TerraformModule) (data.Check, error)) {
			defer wg.Done()

			check, err := checkFunction(modules)
			// Apply whitelist on Terraform code blocks checks errors
			// This only cover Terraform resources that have a POS attribute
			for i := len(check.Errors) - 1; i >= 0; i-- {
				check, _ = helpers.ApplyWhitelistOnCodeBlock(check, i, modules)
				check, _ = helpers.ApplyWhitelistOnModule(check, i, modules)
			}
			// Replace the check error status with the array after whitelisting
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
