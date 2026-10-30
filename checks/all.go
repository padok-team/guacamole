package checks

import (
	"sync"

	"github.com/padok-team/guacamole/checks/profile"
	"github.com/padok-team/guacamole/checks/state"
	"github.com/padok-team/guacamole/checks/static/tf"
	"github.com/padok-team/guacamole/checks/static/tg"
	"github.com/padok-team/guacamole/data"
)

func All(layers []*data.Layer) []data.Check {
	// List of checks to perform

	var checkResults []data.Check

	checkTypes := 2

	wg := new(sync.WaitGroup)
	wg.Add(checkTypes)

	c := make(chan []data.Check, checkTypes)
	defer close(c)

	go func() {
		defer wg.Done()
		c <- state.StateChecks(layers)
	}()

	go func() {
		defer wg.Done()
		c <- tf.ModuleStaticChecks()
	}()

	wg.Wait()

	for i := 0; i < checkTypes; i++ {
		checkResults = append(checkResults, <-c...)
	}

	return checkResults
}

func ModuleStaticChecks() []data.Check {
	return tf.ModuleStaticChecks()
}

func LayerStaticChecks() []data.Check {
	return tg.LayerStaticChecks()
}

func StateChecks(layers []*data.Layer) []data.Check {
	return state.StateChecks(layers)
}

func Profile(codebase data.Codebase, verbose bool) {
	profile.Profile(codebase, verbose)
}
