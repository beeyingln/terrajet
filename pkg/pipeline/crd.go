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
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	twtypes "github.com/muvaf/typewriter/pkg/types"
	"github.com/muvaf/typewriter/pkg/wrapper"
	"github.com/pkg/errors"

	"github.com/crossplane/terrajet/pkg/config"
	"github.com/crossplane/terrajet/pkg/pipeline/templates"
	tjtypes "github.com/crossplane/terrajet/pkg/types"
)

// GenStatement is printed on every generated file.
const GenStatement = "// Code generated by terrajet. DO NOT EDIT."

// NewCRDGenerator returns a new CRDGenerator.
func NewCRDGenerator(pkg *types.Package, rootDir, providerShortName, group, version string) *CRDGenerator {
	return &CRDGenerator{
		LocalDirectoryPath: filepath.Join(rootDir, "apis", strings.ToLower(strings.Split(group, ".")[0]), version),
		LicenseHeaderPath:  filepath.Join(rootDir, "hack", "boilerplate.go.txt"),
		Group:              group,
		ProviderShortName:  providerShortName,
		pkg:                pkg,
	}
}

// CRDGenerator takes certain information referencing Terraform resource definition
// and writes kubebuilder CRD file.
type CRDGenerator struct {
	LocalDirectoryPath string
	Group              string
	ProviderShortName  string
	LicenseHeaderPath  string

	pkg *types.Package
}

// Generate builds and writes a new CRD out of Terraform resource definition.
func (cg *CRDGenerator) Generate(cfg *config.Resource) (string, error) {
	file := wrapper.NewFile(cg.pkg.Path(), cg.pkg.Name(), templates.CRDTypesTemplate,
		wrapper.WithGenStatement(GenStatement),
		wrapper.WithHeaderPath(cg.LicenseHeaderPath),
	)
	for _, omit := range cfg.ExternalName.OmittedFields {
		delete(cfg.TerraformResource.Schema, omit)
	}

	cfg.TerraformResource.Schema["id"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}

	gen, err := tjtypes.NewBuilder(cg.pkg).Build(cfg)
	if err != nil {
		return "", errors.Wrapf(err, "cannot build types for %s", cfg.Kind)
	}
	// TODO(muvaf): TypePrinter uses the given scope to see if the type exists
	// before printing. We should ideally load the package in file system but
	// loading the local package will result in error if there is
	// any compilation errors, which is the case before running kubebuilder
	// generators. For now, we act like the target package is empty.
	pkg := types.NewPackage(cg.pkg.Path(), cg.pkg.Name())
	typePrinter := twtypes.NewPrinter(file.Imports, pkg.Scope(), twtypes.WithComments(gen.Comments))
	typesStr, err := typePrinter.Print(gen.Types)
	if err != nil {
		return "", errors.Wrap(err, "cannot print the type list")
	}
	vars := map[string]interface{}{
		"Types": typesStr,
		"CRD": map[string]string{
			"APIVersion":      cfg.Version,
			"Group":           cg.Group,
			"Kind":            cfg.Kind,
			"ForProviderType": gen.ForProviderType.Obj().Name(),
			"AtProviderType":  gen.AtProviderType.Obj().Name(),
		},
		"Provider": map[string]string{
			"ShortName": cg.ProviderShortName,
		},
		"XPCommonAPIsPackageAlias": file.Imports.UsePackage(tjtypes.PackagePathXPCommonAPIs),
		"IsAggregateResource":      strings.Compare(cfg.Kind, "Resource") == 0,
	}
	filePath := filepath.Join(cg.LocalDirectoryPath, fmt.Sprintf("zz_%s_types.go", strings.ToLower(cfg.Kind)))
	return gen.ForProviderType.Obj().Name(), errors.Wrap(file.Write(filePath, vars, os.ModePerm), "cannot write crd file")
}
