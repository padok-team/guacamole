package tg

import (
	"os"
	"sync"

	"github.com/padok-team/guacamole/data"
	log "github.com/sirupsen/logrus"

	"golang.org/x/exp/slices"
)

func LayerStaticChecks() []data.Check {
	// Add static checks here
	checks := map[string]func() (data.Check, error){
		"Dry":                 Dry,
		"TerragruntLeafLayer": TerragruntLeafLayer,
		"LocalIsUseful":       LocalIsUseful,
		"UnusedInputs":        UnusedInputs,
		"CodeQualityTg":       CodeQualityTg,
	}

	var checkResults []data.Check

	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for name, checkFunction := range checks {
		go func(name string, checkFunction func() (data.Check, error)) {
			defer wg.Done()
			log.Debugf("[ %s ] Running", name)
			check, err := checkFunction()
			if err != nil {
				log.Errorf("[ %s ] Failed: %v", name, err)
				os.Exit(1)
			}
			log.Debugf("[ %s ] status=%s, errors=%d", name, check.Status, len(check.Errors))
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
