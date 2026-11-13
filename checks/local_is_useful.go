package checks

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

// LocalIsUseful checks that every local of a .hcl file is justified: it is reused
// (DRY) or it abstracts complexity. Otherwise its value should be inlined.
func LocalIsUseful() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_DRY_002",
		Name:              "A local should only be defined if it is reused (DRY) or abstracts complexity",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html",
		Status:            "✅",
	}

	paths, err := findFiles(viper.GetString("codebase-path"), func(name string) bool { return filepath.Ext(name) == ".hcl" })
	if err != nil {
		return dataCheck, err
	}

	files := []*hclLocalsFile{}
	for _, path := range paths {
		file, err := scanLocalsInFile(path)
		if err != nil {
			log.Warnf("[TG_DRY_002] skipping %s: %s", path, err)
		} else if file != nil {
			files = append(files, file)
		}
	}

	for _, file := range files {
		dataCheck.Errors = append(dataCheck.Errors, file.uselessLocals(files)...)
	}

	if len(dataCheck.Errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

// hclLocalsFile holds the locals defined by a .hcl file and the locals it references.
type hclLocalsFile struct {
	path string
	defs []*hclsyntax.Attribute
	// refs counts the `local.<name>` references, externalRefs the `<alias>.locals.<name>` ones.
	refs, externalRefs map[string]int
}

// scanLocalsInFile returns nil when the file neither defines nor reads any local.
func scanLocalsInFile(path string) (*hclLocalsFile, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Cheap filter: avoid parsing files that cannot contain a local.
	if !bytes.Contains(src, []byte("locals")) {
		return nil, nil
	}

	hclFile, diags := hclparse.NewParser().ParseHCL(src, path)
	if diags.HasErrors() {
		return nil, diags
	}

	body := hclFile.Body.(*hclsyntax.Body)
	file := &hclLocalsFile{path: path, refs: map[string]int{}, externalRefs: map[string]int{}}

	for _, block := range body.Blocks {
		if block.Type == "locals" {
			file.defs = slices.AppendSeq(file.defs, maps.Values(block.Body.Attributes))
		}
	}
	// Map iteration is random: keep the report stable.
	slices.SortFunc(file.defs, func(a, b *hclsyntax.Attribute) int { return a.SrcRange.Start.Line - b.SrcRange.Start.Line })

	hclsyntax.VisitAll(body, func(node hclsyntax.Node) hcl.Diagnostics {
		switch expr := node.(type) {
		case *hclsyntax.ScopeTraversalExpr: // local.x, include.root.locals.x
			steps := attrSteps(expr.Traversal[1:])
			if expr.Traversal.RootName() == "local" && len(steps) > 0 {
				file.refs[steps[0]]++
			}
			file.countExternalRef(steps)
		case *hclsyntax.RelativeTraversalExpr: // read_terragrunt_config("root.hcl").locals.x
			file.countExternalRef(attrSteps(expr.Traversal))
		}
		return nil
	})

	return file, nil
}

// countExternalRef counts the local read after a "locals" step.
func (f *hclLocalsFile) countExternalRef(steps []string) {
	if i := slices.Index(steps, "locals"); i >= 0 && i+1 < len(steps) {
		f.externalRefs[steps[i+1]]++
	}
}

// uselessLocals returns an error for every local that is neither reused nor complex.
func (f *hclLocalsFile) uselessLocals(all []*hclLocalsFile) []data.Error {
	errs := []data.Error{}

	for _, def := range f.defs {
		used := f.refs[def.Name] + f.externalUsage(def.Name, all)
		if used > 1 || abstractsComplexity(def.Expr) {
			continue
		}

		errs = append(errs, data.Error{
			Path:       f.path,
			LineNumber: def.SrcRange.Start.Line,
			Description: fmt.Sprintf(
				"local \"%s\" is used %d time(s) and does not abstract complexity: inline its value instead of defining a local",
				def.Name, used,
			),
		})
	}

	return errs
}

// externalUsage counts the reads of the local in the other files below this one (children read their parents).
// ponytail: matched by name only, a same-named local read elsewhere below can hide an unused one.
func (f *hclLocalsFile) externalUsage(name string, all []*hclLocalsFile) int {
	total := 0

	for _, other := range all {
		if rel, err := filepath.Rel(filepath.Dir(f.path), other.path); other != f && err == nil && !strings.HasPrefix(rel, "..") {
			total += other.externalRefs[name]
		}
	}

	return total
}

// attrSteps returns the names of the leading attribute steps of a traversal.
func attrSteps(traversal hcl.Traversal) []string {
	steps := []string{}
	for _, step := range traversal {
		attr, ok := step.(hcl.TraverseAttr)
		if !ok {
			break
		}
		steps = append(steps, attr.Name)
	}

	return steps
}

// abstractsComplexity reports whether the value combines several references or calls a function.
func abstractsComplexity(expr hclsyntax.Expression) bool {
	distinct := map[string]bool{}
	for _, traversal := range expr.Variables() {
		distinct[traversal.RootName()+"."+strings.Join(attrSteps(traversal[1:]), ".")] = true
	}

	hasFunctionCall := false
	hclsyntax.VisitAll(expr, func(node hclsyntax.Node) hcl.Diagnostics {
		_, isCall := node.(*hclsyntax.FunctionCallExpr)
		hasFunctionCall = hasFunctionCall || isCall
		return nil
	})

	return len(distinct) > 1 || hasFunctionCall
}
