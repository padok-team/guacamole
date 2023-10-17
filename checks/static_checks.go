package checks

import (
	"guacamole/data"
	"os"
	"strings"
	"sync"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
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

	// Open the YAML file .guacamole.baseline.yaml and check if the checks are enabled or not
	// If the check is enabled, we add it to the checkResults
	baseline, err := os.ReadFile(".guacamole.baseline.yaml")
	if err != nil {
		panic(err)
	}

	baselineParsed := data.Baseline{}

	// Parse the YAML file
	err = yaml.Unmarshal(baseline, &baselineParsed)
	if err != nil {
		panic(err)
	}

	for k, ignored := range baselineParsed.Ignore {
		for i, check := range checkResults {
			if check.ID == k {
				for _, checkResult := range checkResults {
					if checkResult.ID == k {
						for j, checkError := range checkResult.Errors {
							// If ignored string is contained in checkerror string, we remove it from the checkResult.Errors
							for _, ignore := range ignored {
								if strings.Contains(checkError, ignore) {
									if len(checkResults[i].Errors) == 1 {
										checkResults[i].Status = "✅"
										checkResults[i].Errors = []string{}
									} else {
										checkResults[i].Errors = append(checkResult.Errors[:j], checkResult.Errors[j+1:]...)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return checkResults
}
