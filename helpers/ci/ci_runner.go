package ci

import "fmt"

func Run() (int, error) {
	cfg, err := loadConfig()
	if err != nil {
		return 0, err
	}

	layerDirs, moduleDirs, err := detectDirs(cfg)
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

	body := buildCommentBody(results, overallScore, totals.overallPass, totals.overallTotal)

	if cfg.platform == platformGithub {
		writeGithubStepOutputs(body, overallScore, totals.overallPass, totals.overallTotal)
	}

	if cfg.postComment {
		post := postGitlabComment
		if cfg.platform == platformGithub {
			post = postGithubComment
		}
		if err := post(body); err != nil {
			return 0, err
		}
	}

	if totals.hasError && cfg.failOnError {
		return 1, nil
	}

	return 0, nil
}
