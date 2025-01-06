package helpers

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

func GetIgnoreingComments(path string) ([]data.IgnoreComment, error) {
	ignoreComments := []data.IgnoreComment{}
	// Parse the file to get ignore comments
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Read the file and find comments containing guacamole-ignore
	scanner := bufio.NewScanner(file)
	i := 1 //Set cursor to the start of the file
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "guacamole-ignore") {
			ignoreComment := data.IgnoreComment{}
			// Regex to match the check ID in the form of TF/TG_XXX_0XX
			regexp := regexp.MustCompile(`(T[F|G]_(\w+)_\d+)`)
			match := regexp.FindStringSubmatch(line)
			if len(match) > 0 {
				ignoreComment.CheckID = match[0]
				ignoreComment.LineNumber = i
				ignoreComment.Path = path

				// Attach comment to an object
			}
			ignoreComments = append(ignoreComments, ignoreComment)
		}
		i++
	}
	return ignoreComments, nil
}

// Get a list of ignoreing in the .guacamoleignore file
// To enable ignoreing of providers
// The format of the file should be: path of the module - check ID

func GetIgnoreingInFile() ([]data.IgnoreModule, error) {
	ignorefile := viper.GetString("guacamoleignore-path")
	ignore := []data.IgnoreModule{}
	// Parse the file to get ignore comments
	file, err := os.Open(ignorefile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Read the file and find comments lines module path and check ID
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		path := parts[0]
		id := parts[1]
		// Split by commers to get multiple checks
		for _, id := range strings.Split(id, ",") {
			// Add check to validate the format of the check ID and the path
			ignoreModule := data.IgnoreModule{}
			ignoreModule.CheckID = id
			ignoreModule.ModulePath = path
			ignore = append(ignore, ignoreModule)
		}
	}
	return ignore, nil
}

// AssociateIgnoreingComments associates the ignoreing comments to a resource (Resource, Data, Variable or Output)
// Not provider config
func AssociateIgnoreingComments(ignoreCommentOfFile []data.IgnoreComment, keys []string, resourcesInFile map[string]data.TerraformCodeBlock, modules map[string]data.TerraformModule, path string) {
	for _, ignoreingComment := range ignoreCommentOfFile {
		previousPos := 0
		for _, i := range keys {
			// Find closest code block within the file
			if previousPos < ignoreingComment.LineNumber && ignoreingComment.LineNumber < resourcesInFile[i].Pos {
				r := modules[filepath.Dir(path)].Resources[i]
				r.IgnoreComments = append(r.IgnoreComments, ignoreingComment)
				modules[filepath.Dir(path)].Resources[i] = r
				break
			}
			previousPos = resourcesInFile[i].Pos
		}
	}
}

func AssociateIgnoreingCommentsOnModule(ignoreOnModule []data.IgnoreModule, path string, modules map[string]data.TerraformModule) {
	for _, ignore := range ignoreOnModule {
		if ignore.ModulePath == path {
			module := modules[path]
			module.Ignore = append(module.Ignore, ignore)
			modules[path] = module
		}
	}
}

// Remove the check from the list if it is ignoreed in the code
func ApplyIgnoreOnCodeBlock(checks data.Check, indexOfCheckedcheck int, modules map[string]data.TerraformModule) (data.Check, error) {
	for _, module := range modules {
		for _, resource := range module.Resources {
			for _, ignore := range resource.IgnoreComments {
				if strings.Contains(checks.Errors[indexOfCheckedcheck].Path, resource.FilePath) && checks.Errors[indexOfCheckedcheck].LineNumber == resource.Pos && checks.ID == ignore.CheckID {
					checks.Errors = append(checks.Errors[:indexOfCheckedcheck], checks.Errors[indexOfCheckedcheck+1:]...)
					return checks, nil
				}
			}
		}
	}
	return checks, nil
}

// Remove the check from the list if it is ignoreed at the module level
func ApplyIgnoreOnModule(checks data.Check, indexOfCheckedcheck int, modules map[string]data.TerraformModule) (data.Check, error) {
	for _, module := range modules {
		for _, ignore := range module.Ignore {
			// If check ID and path match, remove the check from the list
			if checks.ID == ignore.CheckID && strings.Contains(checks.Errors[indexOfCheckedcheck].Path, ignore.ModulePath) {
				checks.Errors = append(checks.Errors[:indexOfCheckedcheck], checks.Errors[indexOfCheckedcheck+1:]...)
				return checks, nil
			}
		}
	}
	return checks, nil
}
