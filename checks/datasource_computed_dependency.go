package checks

import (
	"sort"
	"strings"

	"github.com/padok-team/guacamole/data"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// DataSourceComputedDependency checks whether a data source depends on a value
// that can only be known during the apply phase (a managed resource attribute).
//
// When a data source's arguments reference a value that is not known until
// apply (typically an attribute of a managed resource declared in the same
// module), Terraform defers reading the data source to the apply phase. As a
// result, the data source shows up as a "potential change" during the plan,
// which makes plans noisy and harder to trust.
//
// See https://developer.hashicorp.com/terraform/language/data-sources#data-source-behavior
func DataSourceComputedDependency(modules map[string]data.TerraformModule) (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TF_DAT_001",
		Name:              "Data source should not depend on a value computed during apply",
		RelatedGuidelines: "https://developer.hashicorp.com/terraform/language/data-sources#data-source-behavior",
		Status:            "✅",
	}

	for _, module := range modules {
		moduleConf, diags := tfconfig.LoadModule(module.FullPath)
		if diags.HasErrors() {
			return data.Check{}, diags.Err()
		}

		// No data source declared -> nothing to check for this module
		if len(moduleConf.DataResources) == 0 {
			continue
		}

		// Build the set of managed resource addresses ("type.name") of the module.
		// A data source can only be deferred to apply if it references a value
		// computed at apply, which within a module means a managed resource attribute.
		managedResources := make(map[string]bool)
		for _, resource := range moduleConf.ManagedResources {
			managedResources[resource.Type+"."+resource.Name] = true
		}

		// If the module has no managed resource, no data source can depend on a
		// computed value -> OK for this module.
		if len(managedResources) == 0 {
			continue
		}

		// Collect the distinct files that declare a data source.
		files := make(map[string]bool)
		for _, dataResource := range moduleConf.DataResources {
			files[dataResource.Pos.Filename] = true
		}

		parser := hclparse.NewParser()
		for file := range files {
			hclFile, diags := parser.ParseHCLFile(file)
			if diags.HasErrors() {
				return data.Check{}, diags
			}

			body, ok := hclFile.Body.(*hclsyntax.Body)
			if !ok {
				continue
			}

			for _, block := range body.Blocks {
				if block.Type != "data" || len(block.Labels) != 2 {
					continue
				}

				computedRefs := collectManagedResourceRefs(block.Body, managedResources)
				if len(computedRefs) > 0 {
					dataSourceAddress := block.Labels[0] + "." + block.Labels[1]
					dataCheck.Errors = append(dataCheck.Errors, data.Error{
						Path:        file,
						LineNumber:  block.DefRange().Start.Line,
						Description: "data " + dataSourceAddress + " depends on value(s) computed during apply: " + strings.Join(computedRefs, ", "),
					})
				}
			}
		}
	}

	if len(dataCheck.Errors) > 0 {
		dataCheck.Status = "❌"
	}

	return dataCheck, nil
}

// collectManagedResourceRefs walks a data block body (including nested blocks
// such as dynamic filters) and returns the sorted list of managed resource
// addresses ("type.name") referenced in its arguments.
func collectManagedResourceRefs(body *hclsyntax.Body, managedResources map[string]bool) []string {
	found := make(map[string]bool)

	for _, attr := range body.Attributes {
		for _, traversal := range attr.Expr.Variables() {
			if addr := managedResourceAddress(traversal); addr != "" && managedResources[addr] {
				found[addr] = true
			}
		}
	}

	for _, block := range body.Blocks {
		for _, ref := range collectManagedResourceRefs(block.Body, managedResources) {
			found[ref] = true
		}
	}

	refs := make([]string, 0, len(found))
	for ref := range found {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	return refs
}

// managedResourceAddress returns the "type.name" address of a traversal when it
// refers to a managed resource, or an empty string otherwise. References to
// variables, locals, other data sources, modules, and meta-arguments are not
// managed resources and are known at plan time.
func managedResourceAddress(traversal hcl.Traversal) string {
	if traversal.IsRelative() || len(traversal) < 2 {
		return ""
	}

	root := traversal.RootName()
	switch root {
	case "var", "local", "data", "module", "each", "count", "self", "path", "terraform":
		return ""
	}

	// The second element must be the resource name (an attribute access).
	nameStep, ok := traversal[1].(hcl.TraverseAttr)
	if !ok {
		return ""
	}

	return root + "." + nameStep.Name
}
