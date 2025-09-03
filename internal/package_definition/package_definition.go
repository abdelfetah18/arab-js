package package_definition

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type PackageDefinition struct {
	Name        string `yaml:"الاسم"`
	Version     string `yaml:"الاصدار"`
	Description string `yaml:"الوصف"`
	Main        string `yaml:"الملف الرئيسي"`
	Author      string `yaml:"المؤلف"`
	Homepage    string `yaml:"الصفحة الرئيسية"`
}

func (p *PackageDefinition) ToString() string {
	return fmt.Sprintf(`الاسم: %s
الاصدار: "%s"
الوصف: "%s"
الملف الرئيسي: "%s"
المؤلف: "%s"
الصفحة الرئيسية: "%s"
`,
		p.Name,
		p.Version,
		p.Description,
		p.Main,
		p.Author,
		p.Homepage,
	)
}

func FromYAMLFile(filePath string) (*PackageDefinition, error) {
	rawContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pkgDef PackageDefinition
	if err := yaml.Unmarshal(rawContent, &pkgDef); err != nil {
		return nil, err
	}

	return &pkgDef, nil
}

func DefaultPackageDefinitionYAML(projectName string) string {
	return (&PackageDefinition{
		Name:        projectName,
		Version:     "1.0.0",
		Description: "",
		Main:        "الرئيسي.كود",
		Author:      "",
		Homepage:    "",
	}).ToString()
}
