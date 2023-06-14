package main

import (
	"fmt"
	"github.com/highgrav/munchkin/internal/util"
	"quamina.net/go/quamina"
)

func (a *application) newPool() error {
	p := util.NewObjectPool[quamina.Quamina](a.config.poolSize)
	a.logger.Info(fmt.Sprintf("Creating object pool with %d entries...", a.config.poolSize))
	for x := 0; x < a.config.poolSize; x++ {
		p.Pool <- a.matcher.Copy()
	}
	a.pool = p
	return nil
}
