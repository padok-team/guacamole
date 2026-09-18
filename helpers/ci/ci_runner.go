package ci

import "fmt"

func Run() (int, error) {
	cfg, err := loadConfig()
	if err != nil {
		return 0, err
	}

	layerDirs, moduleDirs, err := detectChangedDirs(cfg.projectDirAbs, cfg.baseBranch, cfg.mrSHA)
	if err != nil {
		return 0, err
	}

	scopes := []scopeDirs{
		{scope: scopeLayer, dirs: layerDirs},
		{scope: scopeModule, dirs: moduleDirs},
	}

	results, totals, err := runAllScopeScans(scopes, cfg.projectDirAbs)
	if err != nil {
		return 0, err
	}

	overallScore := scorePercent(totals.overallPass, totals.overallTotal)
	fmt.Printf("CI summary: %d%% (%d/%d)\n", overallScore, totals.overallPass, totals.overallTotal)

	if cfg.postComment {
		if err := postGitlabComment(results, overallScore, totals.overallPass, totals.overallTotal); err != nil {
			return 0, err
		}
	}

	if totals.hasError {
		return 1, nil
	}

	return 0, nil
}
