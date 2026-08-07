package checks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

// LocalIsUseful checks that every local defined in a .hcl file is actually
// worth being a local. A local is only justified if:
//   - it is reused (referenced more than once in the file) -> DRY, or
//   - it abstracts complexity (its value merges several inputs).
//
// A local that is used at most once and whose value is a plain literal or a
// simple alias should be inlined instead of being declared as a local.
func LocalIsUseful() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_DRY_002",
		Name:              "A local should only be defined if it is reused (DRY) or abstracts complexity",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html",
		Status:            "✅",
	}

	codebasePath := viper.GetString("codebase-path")
	errorsFound := []data.Error{}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to scan codebase path: %w", err)
		}

		if info.IsDir() {
			if terragruntCacheRegexp.MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".hcl" || terragruntCacheRegexp.MatchString(path) {
			return nil
		}

		log.Debugf("[TG_DRY_002] scanning %s", path)
		fileErrors, checkErr := checkLocalsInFile(path)
		if checkErr != nil {
			log.Warnf("[TG_DRY_002] skipping %s: %s", path, checkErr)
			return nil
		}
		errorsFound = append(errorsFound, fileErrors...)

		return nil
	})
	if err != nil {
		return dataCheck, err
	}

	dataCheck.Errors = errorsFound
	if len(errorsFound) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

type localDefinition struct {
	name  string
	expr  hclsyntax.Expression
	block hcl.Range
}

// checkLocalsInFile parses a single .hcl file and returns an error for every
// local that is neither reused nor abstracting complexity.
func checkLocalsInFile(path string) ([]data.Error, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Skip files with no locals block without paying the cost of a full HCL parse.
	if !bytes.Contains(src, []byte("locals")) {
		log.Debugf("[TG_DRY_002] no locals keyword in %s, skipping", path)
		return nil, nil
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse %s: %s", path, diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, nil
	}

	// A file can hold several `locals` blocks, gather every declared local.
	localDefs := []localDefinition{}
	for _, block := range body.Blocks {
		if block.Type != "locals" {
			continue
		}
		for name, attr := range block.Body.Attributes {
			localDefs = append(localDefs, localDefinition{
				name:  name,
				expr:  attr.Expr,
				block: attr.SrcRange,
			})
		}
	}

	if len(localDefs) == 0 {
		log.Debugf("[TG_DRY_002] no locals in %s, skipping", path)
		return nil, nil
	}
	log.Debugf("[TG_DRY_002] found %d local(s) in %s", len(localDefs), path)

	// Count how many times each local is referenced across the whole file.
	usage := map[string]int{}
	countLocalReferences(body, usage)

	fileErrors := []data.Error{}
	for _, def := range localDefs {
		reused := usage[def.name] > 1
		abstractsComplexity := mergesMultipleInputs(def.expr)

		log.Debugf("[TG_DRY_002] local %q: used=%d reused=%v abstractsComplexity=%v", def.name, usage[def.name], reused, abstractsComplexity)

		if reused || abstractsComplexity {
			continue
		}

		fileErrors = append(fileErrors, data.Error{
			Path:       path,
			LineNumber: def.block.Start.Line,
			Description: fmt.Sprintf(
				"local \"%s\" is used %d time(s) and does not abstract complexity: inline its value instead of defining a local",
				def.name, usage[def.name],
			),
		})
	}

	return fileErrors, nil
}

// countLocalReferences walks every attribute of the body (recursively into
// nested blocks) and counts the `local.<name>` traversals it references.
func countLocalReferences(body *hclsyntax.Body, usage map[string]int) {
	for _, attr := range body.Attributes {
		for _, traversal := range attr.Expr.Variables() {
			if traversal.RootName() != "local" || len(traversal) < 2 {
				continue
			}
			if step, ok := traversal[1].(hcl.TraverseAttr); ok {
				usage[step.Name]++
			}
		}
	}
	for _, block := range body.Blocks {
		countLocalReferences(block.Body, usage)
	}
}

// mergesMultipleInputs reports whether a local's value abstracts complexity,
// i.e. it combines several inputs. This is true when the expression either
// references more than one distinct traversal or calls a function (merge,
// concat, format, ...).
func mergesMultipleInputs(expr hclsyntax.Expression) bool {
	distinct := map[string]struct{}{}
	for _, traversal := range expr.Variables() {
		distinct[traversalKey(traversal)] = struct{}{}
	}
	if len(distinct) > 1 {
		return true
	}

	hasFunctionCall := false
	hclsyntax.VisitAll(expr, func(node hclsyntax.Node) hcl.Diagnostics {
		if _, ok := node.(*hclsyntax.FunctionCallExpr); ok {
			hasFunctionCall = true
		}
		return nil
	})

	return hasFunctionCall
}

// traversalKey builds a stable key identifying a traversal (e.g. "local.env.region")
// so distinct references can be de-duplicated.
func traversalKey(traversal hcl.Traversal) string {
	key := traversal.RootName()
	for _, step := range traversal[1:] {
		if attrStep, ok := step.(hcl.TraverseAttr); ok {
			key += "." + attrStep.Name
		}
	}
	return key
}
