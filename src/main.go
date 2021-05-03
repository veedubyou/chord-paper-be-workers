package main

import (
	"chord-paper-be-workers/src/application"
)

func main() {
	app := application.NewApp()
	app.Start()
	waitForever()
}

func waitForever() {
	<-make(chan bool)
}
