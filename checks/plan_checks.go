package checks

import (
	"guacamole/data"
	"sync"
)

func PlanChecks(layers []*data.Layer) []data.Check {
	// Add plan checks here
	checks := map[string]func([]*data.Layer) (data.Check, error){
		"IterateUseForEach": IterateUseForEach,
		"RefreshTime":       RefreshTime,
	}

	var checkResults []data.Check

	wg := new(sync.WaitGroup)
	wg.Add(len(checks))

	c := make(chan data.Check, len(checks))
	defer close(c)

	for _, checkFunction := range checks {
		go func(checkFunction func([]*data.Layer) (data.Check, error)) {
			defer wg.Done()

			check, err := checkFunction(layers)
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
