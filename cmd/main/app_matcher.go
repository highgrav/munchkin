package main

import "quamina.net/go/quamina"

func (a *application) newMatcher() error {
	m, err := quamina.New(quamina.WithPatternDeletion(true))
	if err != nil {
		return err
	}
	a.matcher = m
	return nil
}
