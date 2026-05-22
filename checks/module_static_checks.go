package checks

import (
	"os"
	"sync"

	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"
	log "github.com/sirupsen/logrus"

	"golang.org/x/exp/slices"
)

func ModuleStaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func(m map[string]data.TerraformModule) (data.Check, error){
		"TF_MOD_001": RemoteModuleVersion,
		"TF_MOD_002": ProviderInModule,
		"TF_MOD_003": RequiredProviderVersionOperatorInModules,
		"TF_NAM_001": ResourceNamingThisThese,
		"TF_NAM_002": SnakeCase,
		"TF_NAM_003": Stuttering,
		"TF_NAM_004": VarNumberMatchesType,
		"TF_NAM_005": ResourceNaming,
		"TF_VAR_001": VarContainsDescription,
		"TF_VAR_002": VarTypeAny,
		"TF_QUA_001": CodeQualityTf,
	}

	var checkResults []data.Check

	// Find recusively all the modules in the current directory
	modules, err := helpers.GetModules()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for name, checkFunction := range checks {
		go func(name string, checkFunction func(m map[string]data.TerraformModule) (data.Check, error)) {
			defer wg.Done()
			log.Debugf("[ %s ] Running", name)
			check, err := checkFunction(modules)
			if err != nil {
				log.Errorf("[ %s ] Failed: %v", name, err)
				os.Exit(1)
			}
			log.Debugf("[ %s ] status=%s, errors=%d", name, check.Status, len(check.Errors))
			// Apply ignore on Terraform code blocks checks errors
			// This only cover Terraform resources that have a POS attribute
			for i := len(check.Errors) - 1; i >= 0; i-- {
				log.Debugf("   [ %s ] Processing ignore (error index: %d)", check.ID, i)
				originalCheck := check
				check, _ = helpers.ApplyIgnoreOnCodeBlock(check, i, modules)
				if len(originalCheck.Errors) > len(check.Errors) {
					log.Debugf("   [ %s ] Ignored by code block rule", check.ID)
				}

				originalCheck = check
				check, _ = helpers.ApplyIgnoreOnModule(check, i, modules)
				if len(originalCheck.Errors) > len(check.Errors) {
					log.Debugf("   [ %s ] Ignored by module rule", check.ID)
				}
			}
			// Replace the check error status with the array after ignoreing
			if len(check.Errors) == 0 {
				check.Status = "✅"
			}
			c <- check
		}(name, checkFunction)
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
