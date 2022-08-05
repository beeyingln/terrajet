/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/muvaf/typewriter/pkg/wrapper"
	"github.com/pkg/errors"

	"github.com/crossplane/terrajet/pkg/pipeline/templates"
)

// NewTerraformedGenerator returns a new TerraformedGenerator.
func NewToolsGenerator(rootDir, modulePath string) *ToolsGenerator {
	return &ToolsGenerator{
		LocalDirectoryPath: filepath.Join(rootDir, "apis"),
		LicenseHeaderPath:  filepath.Join(rootDir, "hack", "boilerplate.go.txt"),
		ModulePath:         modulePath,
	}
}

// TerraformedGenerator generates conversion methods implementing Terraformed
// interface on CRD structs.
type ToolsGenerator struct {
	LocalDirectoryPath string
	LicenseHeaderPath  string
	ModulePath         string
}

// Generate writes generated Terraformed interface functions
func (atg *ToolsGenerator) Generate(toolsResources map[string]map[string]string) error {
	toolsFile := wrapper.NewFile(filepath.Join(atg.ModulePath, "apis"), "apis", templates.ToolsTemplate,
		wrapper.WithGenStatement(GenStatement),
		wrapper.WithHeaderPath(atg.LicenseHeaderPath),
	)

	resources := make(map[string]string)
	for key, value := range toolsResources {
		if !strings.HasSuffix(key, "_resource") {
			resources[key] = toolsFile.Imports.UsePackage(value["package"]) + value["struct"]
		}
	}

	vars := map[string]interface{}{
		"Resources": resources,
	}

	filePath := filepath.Join(atg.LocalDirectoryPath, "tools", "apis_tools.go")
	return errors.Wrap(toolsFile.Write(filePath, vars, os.ModePerm), "cannot write API tools file")
}
