package helpers

import (
	"fmt"
	"guacamole/data"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func GetModules() ([]data.Module, error) {
	codebasePath := viper.GetString("codebase-path") + "modules/"
	modules := []data.Module{}
	//Get all subdirectories in root path
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
		}
		if info.IsDir() && path != codebasePath {
			modules = append(modules, data.Module{Name: info.Name(), FullPath: path})
		}
		return nil
	})
	return modules, err
}
