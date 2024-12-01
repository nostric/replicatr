package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aviate-labs/agent-go/identity"
	sec "github.com/aviate-labs/secp256k1"
	"github.com/pkg/profile"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/interrupt"
	"realy.lol/lol"

	"github.com/Hubmakerlabs/replicatr/agent"
	"github.com/Hubmakerlabs/replicatr/config"
)

func cmdEnv(cfg *config.C) {
	config.PrintEnv(cfg, os.Stdout)
	os.Exit(0)
}

func cmdPubkey(cfg *config.C) {
	var err E
	var secKeyBytes B
	secKeyBytes, err = hex.Dec(cfg.CanisterSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error decoding canister secret key: %s\n", err)
		return
	}
	var secKey *sec.PrivateKey
	secKey, _ = sec.PrivKeyFromBytes(sec.S256(), secKeyBytes)
	var id *identity.Secp256k1Identity
	id, err = identity.NewSecp256k1Identity(secKey)
	if err != nil {
		log.E.F("Error creating identity: %s\n", err)
		os.Exit(1)
	}
	publicKeyBase64 := base64.StdEncoding.EncodeToString(id.PublicKey())
	fmt.Printf("Your Canister-Facing Relay Pubkey is:\n%s\n", publicKeyBase64)
	os.Exit(0)
}

func cmdAddRelay(cfg *config.C, c context.T) {
	var err E
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr,
			"ERROR: addrelay requires at least 3 arguments:\n\n"+
				"\t%s addrelay <pubkey> <admin: true/false>\n\n"+
				"got: %v", os.Args[0], os.Args[1:])
	}
	var admin bool
	switch strings.ToLower(os.Args[3]) {
	case "true":
		admin = true
	case "false":
	default:
		log.E.F("3rd parameter must be true or false, got: '%s'\n\n%v\n",
			os.Args[3], os.Args)
		os.Exit(1)
	}
	var a *agent.Backend
	a, err = agent.New(c, cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	err = a.AddUser(os.Args[2], admin)
	if err != nil {
		log.E.F("Error adding user: %s\n", err)
		os.Exit(1)
	}
	perm := "user"
	if admin {
		perm = "admin"
	}
	log.I.F("User %s added with %s level access\n", os.Args[2], perm)
	os.Exit(0)
}

func cmdRemoveRelay(cfg *config.C, c context.T) {
	var err E
	var a *agent.Backend
	a, err = agent.New(c, cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	err = a.RemoveUser(os.Args[2])
	if err != nil {
		log.E.F("Error removing user: %s\n", err)
		os.Exit(1)
	}
	log.I.F("User %s removed\n", os.Args[2])
	os.Exit(0)
}

func cmdGetPermission(cfg *config.C, c context.T) {
	var err E
	var a *agent.Backend
	a, err = agent.New(c, cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	var perm S
	perm, err = a.GetPermission()
	if err != nil {
		log.E.F("ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("This relay has %s level access\n", perm)
	os.Exit(0)
}

func main() {
	var err E
	var cfg *config.C
	if cfg, err = config.New(); chk.T(err) || config.HelpRequested() {
		if err != nil {
			log.E.F("ERROR: %s", err)
			os.Exit(1)
		}
		config.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	c, cancel := context.Cancel(context.Bg())
	interrupt.AddHandler(cancel)
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "env":
			cmdEnv(cfg)
		case "pubkey":
			cmdPubkey(cfg)
		case "addrelay":
			cmdAddRelay(cfg, c)
		case "removerelay":
			cmdRemoveRelay(cfg, c)
		case "getpermission":
			cmdGetPermission(cfg, c)
		default:
			log.E.F("ERROR: unknown command '%s'\ncommandline: %v\n", os.Args[1], os.Args)
			os.Exit(1)
		}
	}
	log.I.Ln("log level", cfg.LogLevel)
	lol.SetLogLevel(cfg.LogLevel)
	if cfg.Pprof {
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			chk.E(http.ListenAndServe("127.0.0.1:6060", nil))
		}()
	}
	debug.SetMemoryLimit(int64(cfg.MemLimit))

}
