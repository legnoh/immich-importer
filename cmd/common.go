package cmd

type GlobalFlags struct {
	Debug bool `name:"debug" env:"DEBUG" help:"Enable debug logging."`
}
