package checks

import (
	"context"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terragrunt/pkg/config"
	"github.com/gruntwork-io/terragrunt/pkg/log"
	"github.com/gruntwork-io/terragrunt/pkg/log/format"
	"github.com/gruntwork-io/terragrunt/pkg/options"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/padok-team/guacamole/data"
	"github.com/padok-team/guacamole/helpers"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zclconf/go-cty/cty"
)

func Dry() (data.Check, error) {
	dataCheck := data.Check{
		ID:                "TG_DRY_001",
		Name:              "No duplicate inputs within a layer",
		RelatedGuidelines: "https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html#%EF%B8%8F-context",
		Status:            "✅",
	}
	codebasePath := viper.GetString("codebase-path")
	// Get all layers
	layers, err := findLayers(codebasePath)
	if err != nil {
		fmt.Println("Couldn't find files on codebase-path, is it a terragrunt repo root ?")
		fmt.Println(err)
		return dataCheck, err
	}

	duplicates := []data.Error{}
	for _, layer := range layers {
		// Get all files that are included with Terragrunt include block
		files, err := findFilesInLayers(layer)
		if err != nil {
			logrus.Warnf("[TG_DRY_001] skipping layer %s: %s", layer, err)
			continue
		}
		// If there is more that 1 file, aka if we at least include 1 othe file
		if len(files) > 1 {
			// Get every 2 combination of file possible
			for combination := range helpers.CombinationsStr(files, 2) {
				findings, err := findDuplicationInInputs(combination[0], combination[1])
				if err != nil {
					return dataCheck, err
				}
				for key, f := range findings {
					duplicates = append(duplicates, data.Error{
						Path:        combination[0] + " and " + combination[1],
						LineNumber:  -1,
						Description: fmt.Sprintf("input %q has the same value %q in both files", key, f),
					})
				}
			}
		}
	}
	dataCheck.Errors = duplicates

	if len(duplicates) > 0 {
		dataCheck.Status = "❌"
	}
	return dataCheck, nil
}

// partialParseLayer lets Terragrunt partially parse a layer (locals, functions, includes) plus the given sections.
func partialParseLayer(path string, decode ...config.PartialDecodeSectionType) (*config.TerragruntConfig, error) {
	// Terragrunt functions like get_path_to_repo_root() need an absolute path.
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// Create new terragrunt option scoped to the file we are scanning
	options, _ := options.NewTerragruntOptionsWithConfigPath(path)
	options.OriginalTerragruntConfigPath = path
	formatter := format.NewFormatter(format.NewKeyValueFormatPlaceholders())
	formatter.SetDisabledColors(true)
	logger := log.New(log.WithFormatter(formatter), log.WithOutput(io.Discard)) // errors are returned, not printed
	ctx, configContext := config.NewParsingContext(context.Background(), logger, options)
	return config.PartialParseConfigFile(ctx, configContext.WithDecodeList(decode...), logger, path, nil)
}

func findFilesInLayers(path string) ([]string, error) {
	files := []string{}
	terragruntConfig, err := partialParseLayer(path)
	if err != nil {
		return files, err
	}
	// Add initial terragrunt.hcl file
	files = append(files, path)
	// Add all includes files
	for _, i := range terragruntConfig.ProcessedIncludes {
		if strings.HasPrefix(i.Path, ".") || !strings.HasPrefix(i.Path, "/") {
			// Convert relative path to absolute
			files = append(files, filepath.Clean(filepath.Dir(path)+"/"+i.Path))
		} else {
			files = append(files, i.Path)
		}
	}
	return files, nil
}

// We use a custom struct and not the one from Terragrunt because it's simplier
type Input struct {
	Inputs *cty.Value `hcl:"inputs,attr"`
}

func findDuplicationInInputs(file1 string, file2 string) (map[string]string, error) {
	hcl := hclparse.NewParser()
	var input Input
	hclFile, err := hcl.ParseHCLFile(file1)
	if err != nil {
		return nil, err
	}
	gohcl.DecodeBody(hclFile.Body, nil, &input)
	var input2 Input
	hclfile2, err := hcl.ParseHCLFile(file2)
	if err != nil {
		return nil, err
	}
	gohcl.DecodeBody(hclfile2.Body, nil, &input2)
	// Recursive function to compare two hcl file and check if their are duplication of inputs
	if input.Inputs != nil && input2.Inputs != nil {
		findings := recursiveLookupHclFileToFindDuplication(*input.Inputs, *input2.Inputs, "")
		if len(findings) > 0 {
			return findings, nil
		}
	}
	return nil, nil
}

// Recursive function that checks two cty.Value from a parsed hcl file, ctx evolve by going down a map
func recursiveLookupHclFileToFindDuplication(file1 cty.Value, file2 cty.Value, ctx string) map[string]string {
	findings := make(map[string]string)
	for keyf1, i := range file1.AsValueMap() {
		for keyf2, j := range file2.AsValueMap() {
			if i.Type() == cty.String && j.Type() == cty.String {
				if i == j && keyf1 == keyf2 { // If 2 strings or number are the same?
					findings[strings.Trim(ctx+"."+keyf1, ".")] = i.AsString()
				}
			} else if i.Type() == cty.Number && j.Type() == cty.Number {
				// Numbers are arbitrary-precision decimal numbers
				// https://pkg.go.dev/github.com/zclconf/go-cty/cty#Type
				// We convert them with AsBigFloat().String()
				if i.AsBigFloat().String() == j.AsBigFloat().String() && keyf1 == keyf2 {
					findings[strings.Trim(ctx+"."+keyf1, ".")] = i.AsBigFloat().String()
				}
			} else if i.Type().IsObjectType() && j.Type().IsObjectType() {
				maps.Copy(findings, recursiveLookupHclFileToFindDuplication(i, j, ctx+"."+keyf1))
			}
		}
	}
	return findings
}
