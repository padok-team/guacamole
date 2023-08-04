package checks

import (
	"guacamole/data"
	"sync"
)

func All(layers []data.Layer) []data.Check {
	// List of checks to perform

	var checkResults []data.Check

	checkTypes := 2

	wg := new(sync.WaitGroup)
	wg.Add(checkTypes)

	c := make(chan []data.Check, checkTypes)
	defer close(c)

	go func() {
		defer wg.Done()
		checks := PlanChecks(layers)
		c <- checks
	}()

	go func() {
		defer wg.Done()
		checks := StaticChecks()
		c <- checks
	}()

	wg.Wait()

	for i := 0; i < checkTypes; i++ {
		checkResults = append(checkResults, <-c...)
	}

	return checkResults
}
