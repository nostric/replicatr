package ic

import (
	"github.com/Hubmakerlabs/replicatr/config"
	. "github.com/Hubmakerlabs/replicatr/eventstores"
	"github.com/Hubmakerlabs/replicatr/eventstores/badger"
	"github.com/Hubmakerlabs/replicatr/eventstores/iconly"
	"realy.lol/layer2"
	"realy.lol/store"
)

func New(cfg *config.C) (sto store.I, err er) {
	var l1, l2 store.I
	if l1, err = badger.New(cfg); chk.E(err) {
		return
	}
	if l2, err = iconly.New(cfg); chk.E(err) {
		return
	}
	sto = &layer2.Backend{
		Ctx:           C,
		WG:            Wg,
		L1:            l1,
		L2:            l2,
		PollFrequency: cfg.PollFrequency,
		PollOverlap:   cfg.PollOverlap,
		EventSignal:   nil,
	}
	return
}
