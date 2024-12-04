package iconly

import (
	"github.com/Hubmakerlabs/replicatr/config"
	"github.com/Hubmakerlabs/replicatr/icp"
	"realy.lol/store"
)

func New(cfg *config.C) (sto store.I, err er) {
	log.I.Ln("initializing IC backend")
	if cfg.CanisterAddr == "" || cfg.CanisterId == "" || cfg.CanisterSecret == "" {
		err = errorf.E("missing required canister parameters, got addr: \"%s\" and id: \"%s\"",
			cfg.CanisterAddr, cfg.CanisterId)
		return
	}
	if sto, err = icp.New(cfg.CanisterId, cfg.CanisterAddr, cfg.CanisterSecret); chk.E(err) {
		return
	}
	return
}
