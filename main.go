package main

import (
	"arab_js/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
