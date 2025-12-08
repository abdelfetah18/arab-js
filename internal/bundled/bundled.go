package bundled

import (
	"embed"
)

//go:embed assets/libs/**
var content embed.FS

type LibName = string

const (
	LibNameES5 LibName = "es5"
	LibNameDom LibName = "dom"
)

func ReadLibFile(name LibName) string {
	switch name {
	case "dom":
		data, err := content.ReadFile("assets/libs/lib.dom.d.arabjs")
		if err != nil {
			panic(err)
		}
		return string(data)
	case "es5":
		data, err := content.ReadFile("assets/libs/lib.es5.d.arts")
		if err != nil {
			panic(err)
		}
		return string(data)
	}

	panic("No lib found '" + name + "'")
}
