package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/Hubmakerlabs/replicatr/icp"
	"github.com/aviate-labs/agent-go/identity"
	sec "github.com/aviate-labs/secp256k1"
	"realy.lol/hex"

	"github.com/Hubmakerlabs/replicatr/config"
)

func cmdEnv(cfg *config.C) {
	config.PrintEnv(cfg, os.Stdout)
	os.Exit(0)
}

func cmdPubkey(cfg *config.C) {
	var err er
	var secKeyBytes by
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

func cmdAddRelay(cfg *config.C, c cx, args []st) {
	var err er
	if len(args) < 4 {
		log.E.F("ERROR: addrelay requires at least 3 arguments:\n\n"+
			"\t%s addrelay <pubkey> <admin: true/false>\n\n"+
			"got: %v", args[0], args[1:])
	}
	var admin bool
	switch strings.ToLower(args[3]) {
	case "true":
		admin = true
	case "false":
	default:
		log.E.F("3rd parameter must be true or false, got: '%s'\n\n%v\n",
			args[3], args)
		os.Exit(1)
	}
	var a *icp.T
	a, err = icp.New(cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	err = a.AddUser(args[2], admin)
	if err != nil {
		log.E.F("Error adding user: %s\n", err)
		os.Exit(1)
	}
	perm := "user"
	if admin {
		perm = "admin"
	}
	log.I.F("User %s added with %s level access\n", args[2], perm)
	os.Exit(0)
}

func cmdRemoveRelay(cfg *config.C, c cx, args []st) {
	var err er
	var a *icp.T
	a, err = icp.New(cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	err = a.RemoveUser(args[2])
	if err != nil {
		log.E.F("Error removing user: %s\n", err)
		os.Exit(1)
	}
	log.I.F("User %s removed\n", args[2])
	os.Exit(0)
}

func cmdGetPermission(cfg *config.C, c cx, args []st) {
	var err er
	var a *icp.T
	a, err = icp.New(cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret)
	if err != nil {
		log.E.F("Error creating agent: %s\n", err)
		os.Exit(1)
	}
	var perm st
	perm, err = a.GetPermission()
	if err != nil {
		log.E.F("ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("This relay has %s level access\n", perm)
	os.Exit(0)
}

func runCmds(cfg *config.C, c cx, args []st) {
	if len(args) > 1 {
		switch strings.ToLower(args[1]) {
		case "env":
			cmdEnv(cfg)
		case "pubkey":
			cmdPubkey(cfg)
		case "addrelay":
			cmdAddRelay(cfg, c, args)
		case "removerelay":
			cmdRemoveRelay(cfg, c, args)
		case "getpermission":
			cmdGetPermission(cfg, c, args)
		default:
			log.E.F("ERROR: unknown command '%s'\ncommandline: %v\n", args[1], args)
			os.Exit(1)
		}
	}
}
