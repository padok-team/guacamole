package data

import (
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

type TerraformModule struct {
	Name         string
	FullPath     string
	ModuleConfig tfconfig.Module
	Resources    map[string]TerraformCodeBlock
	Ignore       []IgnoreModule
}

type TerraformCodeBlock struct {
	Name           string
	ModulePath     string
	Pos            int
	FilePath       string
	IgnoreComments []IgnoreComment
}
