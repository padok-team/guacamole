package helpers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/padok-team/guacamole/data"
	"github.com/spf13/viper"
)

func GetWhitelistingComments(path string) ([]data.WhitelistComment, error) {
	whitelistComments := []data.WhitelistComment{}
	// Parse the file to get whitelist comments
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
			whitelistComment := data.WhitelistComment{}
			// Regex to match the check ID in the form of TF/TG_XXX_0XX
			regexp := regexp.MustCompile(`(T[F|G]_(\w+)_\d+)`)
			match := regexp.FindStringSubmatch(line)
			if len(match) > 0 {
				whitelistComment.CheckID = match[0]
				whitelistComment.LineNumber = i
				whitelistComment.Path = path

				// Attach comment to an object
			}
			whitelistComments = append(whitelistComments, whitelistComment)
		}
		i++
	}
	return whitelistComments, nil
}

// Get a list of whitelisting in the .guacamoleignore file
// To enable whitelisting of providers
// The format of the file should be: path of the module - check ID

func GetWhitelistingInFile() ([]data.WhitelistModule, error) {
	whitelistfile := viper.GetString("whitelistfile-path")
	whitelist := []data.WhitelistModule{}
	// Parse the file to get whitelist comments
	file, err := os.Open(whitelistfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Read the file and find comments lines module path and check ID
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			fmt.Println("Invalid line format:", line)
			continue
		}
		path := parts[0]
		id := parts[1]
		// Add check to validate the format of the check ID and the path
		if len(id) > 0 && len(path) > 0 {
			whitelistModule := data.WhitelistModule{}
			whitelistModule.CheckID = id
			whitelistModule.ModulePath = path
			whitelist = append(whitelist, whitelistModule)
		}
	}
	return whitelist, nil
}

// AssociateWhitelistingComments associates the whitelisting comments to a resource (Resource, Data, Variable or Output)
// Not provider config
func AssociateWhitelistingComments(whitelistCommentOfFile []data.WhitelistComment, keys []string, resourcesInFile map[string]data.TerraformCodeBlock, modules map[string]data.TerraformModule, path string) {
	for _, whitelistingComment := range whitelistCommentOfFile {
		previousPos := 0
		for _, i := range keys {
			// Find closest code block within the file
			if previousPos < whitelistingComment.LineNumber && whitelistingComment.LineNumber < resourcesInFile[i].Pos {
				r := modules[filepath.Dir(path)].Resources[i]
				r.WhitelistComments = append(r.WhitelistComments, whitelistingComment)
				modules[filepath.Dir(path)].Resources[i] = r
				break
			}
			previousPos = resourcesInFile[i].Pos
		}
	}
}

func AssociateWhitelistingCommentsOnModule(whitelistOnModule []data.WhitelistModule, path string, modules map[string]data.TerraformModule) {
	for _, whitelist := range whitelistOnModule {
		if whitelist.ModulePath == path {
			module := modules[path]
			module.Whitelist = append(module.Whitelist, whitelist)
			modules[path] = module
			break
		}
	}
}

// Remove the check from the list if it is whitelisted in the code
func ApplyWhitelistOnCodeBlock(checks data.Check, indexOfCheckedcheck int, modules map[string]data.TerraformModule) (data.Check, error) {
	for _, module := range modules {
		for _, resource := range module.Resources {
			for _, whitelist := range resource.WhitelistComments {
				if strings.Contains(checks.Errors[indexOfCheckedcheck].Path, resource.FilePath) && checks.Errors[indexOfCheckedcheck].LineNumber == resource.Pos && checks.ID == whitelist.CheckID {
					checks.Errors = append(checks.Errors[:indexOfCheckedcheck], checks.Errors[indexOfCheckedcheck+1:]...)
					return checks, nil
				}
			}
		}
	}
	return checks, nil
}

// Remove the check from the list if it is whitelisted at the module level
func ApplyWhitelistOnModule(checks data.Check, indexOfCheckedcheck int, modules map[string]data.TerraformModule) (data.Check, error) {
	for _, module := range modules {
		for _, whitelist := range module.Whitelist {
			// If check ID and path match, remove the check from the list
			if checks.ID == whitelist.CheckID && checks.Errors[indexOfCheckedcheck].Path == whitelist.ModulePath {
				checks.Errors = append(checks.Errors[:indexOfCheckedcheck], checks.Errors[indexOfCheckedcheck+1:]...)
				return checks, nil
			}
		}
	}
	return checks, nil
}
