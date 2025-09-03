package bundled

import (
	"embed"
)

//go:embed assets/libs/**
var content embed.FS

func ReadFile(name string) string {
	data, err := content.ReadFile("assets/libs/lib.dom.d.arabjs")
	if err != nil {
		panic(err)
	}

	return string(data)
}
