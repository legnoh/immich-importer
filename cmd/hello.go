package cmd

import (
	"fmt"

	"github.com/legnoh/immich-importer/internal/logger"
)

type HelloCmd struct {
	Name string `arg:"" env:"MY_NAME" name:"name" help:"Name to greet." default:"world"`
}

func (c *HelloCmd) Run(g GlobalFlags) error {
	log := logger.Default

	if c.Name == "devil" {
		log.Debug("devil is not a good name, but we'll greet you anyway...")
		return fmt.Errorf("devil is unwelcomed here...")
	}

	log.Info("greeting user", "name", c.Name)
	fmt.Printf("hello %s!\n", c.Name)

	return nil
}
