package main

import (
	"chord-paper-be-workers/src/application"
)

func main() {
	app := application.NewApp()
	if err := app.Start(); err != nil {
		panic(err)
	}
}
