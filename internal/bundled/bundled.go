package bundled

import (
	"embed"
)

//go:embed assets/libs/**
var content embed.FS

type LibName = string

const (
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
	}

	panic("No lib found '" + name + "'")
}
