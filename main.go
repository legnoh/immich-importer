package main

import (
	"log/slog"

	"github.com/alecthomas/kong"
	"github.com/legnoh/immich-importer/cmd"
	"github.com/legnoh/immich-importer/internal/logger"
)

var version = ""
var appname = "immich-importer"           // CHANGE ME
var appdesc = "Import files into Immich." // CHANGE ME

var cli struct {
	GlobalFlags cmd.GlobalFlags  `embed:""`
	Version     kong.VersionFlag `name:"version" help:"Show version."`

	// Subcommands
	Hello  cmd.HelloCmd  `cmd:"hello" help:"Say hello."`
	Upload cmd.UploadCmd `cmd:"upload" short:"u" help:"Upload files"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name(appname),
		kong.Description(appdesc),
		kong.UsageOnError(),
		kong.Vars{"version": version},
		kong.Bind(cli.GlobalFlags),
	)

	level := slog.LevelInfo
	if cli.GlobalFlags.Debug {
		level = slog.LevelDebug
	}

	log := logger.NewWithLevel(level)
	logger.Default = log
	log.Debug("starting CLI")

	err := ctx.Run()
	if err != nil {
		log.Error("error running command", "msg", err)
		ctx.Exit(1)
	}
}
