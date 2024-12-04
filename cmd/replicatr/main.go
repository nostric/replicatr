package main

import (
	"net/http"
	"os"
	"runtime/debug"

	"github.com/Hubmakerlabs/replicatr/config"
	"github.com/Hubmakerlabs/replicatr/eventstores/badger"
	"github.com/Hubmakerlabs/replicatr/eventstores/ic"
	"github.com/pkg/profile"
	"realy.lol/context"
	"realy.lol/interrupt"
	"realy.lol/lol"
	"realy.lol/store"
)

func main() {
	var err er
	var cfg *config.C
	if cfg, err = config.New(); chk.T(err) || config.HelpRequested() {
		if err != nil {
			log.E.F("ERROR: %s", err)
			os.Exit(1)
		}
		config.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	log.I.S(cfg)
	c, cancel := context.Cancel(context.Bg())
	// if user has specified a run command, execute them
	runCmds(cfg, c, os.Args)
	interrupt.AddHandler(cancel)
	log.I.Ln("log level", cfg.LogLevel)
	lol.SetLogLevel(cfg.LogLevel)
	if cfg.Pprof {
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			chk.E(http.ListenAndServe("127.0.0.1:6060", nil))
		}()
	}
	debug.SetMemoryLimit(int64(cfg.MemLimit))
	// set up the database(s)
	var sto store.I
	switch cfg.EventStore {
	case "badger":
		if sto, err = badger.New(cfg); chk.E(err) {
			os.Exit(1)
			return
		}
	case "ic":
		if sto, err = ic.New(cfg); chk.E(err) {
			os.Exit(1)
			return
		}
	}
	_ = sto
}
