package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kivutar/goro/app"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/updater"
)

func main() {
	if updater.HandleFlags(os.Args[1:], os.Stdout) {
		return
	}
	cfg, err := config.LoadConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	closeLog, err := glog.Configure(cfg.Log)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer closeLog()

	if cfg.UILayout.DumpPath != "" {
		if err := app.DumpUILayout(cfg.UILayout.DumpPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if updater.Enabled(os.Args[1:]) {
		res, err := updater.CheckOnLaunch(os.Args[1:], glog.Infof)
		switch {
		case err != nil:
			// A failed update never blocks play: the installed binary is unchanged.
			glog.Warnf("update check failed: %v", err)
		case res.Outcome == updater.Applied:
			glog.Infof("updated to %s; restarting", res.Version)
			closeLog()
			os.Exit(0)
		case res.Outcome == updater.UpToDate:
			glog.Infof("update up to date %s (build %d)", res.Version, res.Build)
		case res.Reason != "":
			glog.Infof("update %s: %s", res.Outcome, res.Reason)
		}
	}

	game, err := app.New(cfg)
	if err != nil {
		glog.Fatalf("%v", err)
	}

	if cfg.Headless {
		err = render.RunHeadless(context.Background(), game, cfg.Window)
	} else {
		err = render.Run(game, cfg.Window, cfg.Render)
	}
	if err != nil {
		glog.Fatalf("%v", err)
	}
}
