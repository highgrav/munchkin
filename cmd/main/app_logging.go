package main

import (
	"go.uber.org/zap"
	"log"
)

func (app *application) setUpLogging() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	app.logger = l
}

func (a *application) newLogger() error {
	a.setUpLogging()
	return nil
}
