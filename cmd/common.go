package cmd

type GlobalFlags struct {
	Debug          bool   `name:"debug" env:"DEBUG" help:"Enable debug logging."`
	ImmichEndpoint string `help:"Immich endpoint URL." env:"IMC_ENDPOINT" default:"http://localhost:2283/"`
	ImmichApiKey   string `help:"Immich API key." env:"IMC_API_KEY" default:""`
}
