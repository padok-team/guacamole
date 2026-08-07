package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/padok-team/guacamole/checks"
	"github.com/padok-team/guacamole/data"
	helperspkg "github.com/padok-team/guacamole/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func runAllScopeScans(scopes []scopeDirs, projectDir string) ([]result, totals, error) {
	results := make([]result, 0)
	agg := totals{}

	for _, s := range scopes {
		log.WithFields(log.Fields{"scope": s.scope, "count": len(s.dirs)}).Debug("Scanning scope directories")
		for _, dir := range s.dirs {
			result, scanned, err := runScopeScan(s.scope, dir, projectDir)
			if err != nil {
				return nil, totals{}, logAndReturnError(err)
			}
			if !scanned {
				continue
			}

			results = append(results, result)
			agg.overallPass += result.passed
			agg.overallTotal += result.total
			agg.hasError = agg.hasError || result.hasError
		}
	}

	return results, agg, nil
}

func runScopeScan(scope scope, relDir, projectDir string) (result, bool, error) {
	absDir := filepath.Join(projectDir, filepath.FromSlash(relDir))
	if _, err := os.Stat(absDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Skipping missing directory:", absDir)
			return result{}, false, nil
		}
		return result{}, false, logAndReturnErrorf("failed to stat path %q: %w", absDir, err)
	}

	fmt.Printf("Running static %s checks on %s\n", scope, absDir)
	log.WithFields(log.Fields{"scope": scope, "path": absDir}).Debug("Running scoped static checks")

	previousCodebasePath := viper.GetString("codebase-path")
	viper.Set("codebase-path", absDir)
	defer viper.Set("codebase-path", previousCodebasePath)

	var checksResults []data.Check
	switch scope {
	case scopeLayer:
		layerChecks := checks.LayerStaticChecks()
		helperspkg.RenderChecks(layerChecks, true)
		checksResults = layerChecks
	case scopeModule:
		moduleChecks := checks.ModuleStaticChecks()
		helperspkg.RenderChecks(moduleChecks, true)
		checksResults = moduleChecks
	default:
		return result{}, false, logAndReturnErrorf("unsupported CI scope %s", scope)
	}

	passed, total, failing := summarizeChecks(checksResults)
	score := "n/a"
	if total > 0 {
		score = fmt.Sprintf("%d%%", scorePercent(passed, total))
	}

	failingText := "-"
	if len(failing) > 0 {
		failingText = strings.Join(failing, " <br> ")
	}

	return result{
		scope:       scope,
		path:        relDir,
		passed:      passed,
		total:       total,
		score:       score,
		failingText: failingText,
		hasError:    len(failing) > 0,
	}, true, nil
}

func summarizeChecks(checkResults []data.Check) (int, int, []string) {
	passed := 0
	failing := []string{}

	for _, c := range checkResults {
		if c.Status == "✅" {
			passed++
			continue
		}
		if c.Status == "❌" {
			failing = append(failing, fmt.Sprintf("❌ %s - %s", c.ID, c.Name))
		}
	}

	return passed, len(checkResults), failing
}

func scorePercent(passed, total int) int {
	if total == 0 {
		return 0
	}
	return passed * 100 / total
}
