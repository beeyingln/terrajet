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

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/muvaf/typewriter/pkg/wrapper"
	"github.com/pkg/errors"

	"github.com/crossplane/terrajet/pkg/pipeline/templates"
)

// NewClientGenerator returns a new ClientGenerator.
func NewClientGenerator(rootDir, modulePath string) *ClientGenerator {
	return &ClientGenerator{
		LocalDirectoryPath: filepath.Join(rootDir, "internal"),
		LicenseHeaderPath:  filepath.Join(rootDir, "hack", "boilerplate.go.txt"),
		ModulePath:         modulePath,
	}
}

// ClientGenerator generates conversion methods implementing Terraformed
// interface on CRD structs.
type ClientGenerator struct {
	LocalDirectoryPath string
	LicenseHeaderPath  string
	ModulePath         string
}

// Generate writes generated Terraformed interface functions
func (cg *ClientGenerator) Generate(attributes map[string]*tfjson.SchemaAttribute) error {
	clientFile := wrapper.NewFile(filepath.Join(cg.ModulePath, "internal"), "clients", templates.ClientTemplate,
		wrapper.WithGenStatement(GenStatement),
		wrapper.WithHeaderPath(cg.LicenseHeaderPath),
	)

	var attrs = make([]string, 0)
	var sensitiveAttrs = make([]string, 0)
	for key, value := range attributes {
		if value.Sensitive {
			sensitiveAttrs = append(sensitiveAttrs, key)
		} else {
			attrs = append(attrs, key)
		}
	}
	vars := map[string]interface{}{
		"Attributes":          attrs,
		"SensitiveAttributes": sensitiveAttrs,
		"ModulePath":          cg.ModulePath,
	}

	filePath := filepath.Join(cg.LocalDirectoryPath, "clients", "terraform_client.go")
	return errors.Wrap(clientFile.Write(filePath, vars, os.ModePerm), "cannot write Terraform Client file")
}
