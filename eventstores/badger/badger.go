package badger

import (
	"time"

	"github.com/Hubmakerlabs/replicatr/config"
	. "github.com/Hubmakerlabs/replicatr/eventstores"
	"realy.lol/lol"
	"realy.lol/ratel"
	"realy.lol/store"
	"realy.lol/units"
)

func New(cfg *config.C) (sto store.I, err er) {
	sto = ratel.New(
		ratel.BackendParams{
			Ctx:            C,
			WG:             Wg,
			BlockCacheSize: units.Gb * 16,
			LogLevel:       lol.GetLogLevel(cfg.DbLogLevel),
			MaxLimit:       ratel.DefaultMaxLimit,
			Extra: []no{
				cfg.DBSizeLimit,
				cfg.DBLowWater,
				cfg.DBHighWater,
				cfg.GCFrequency * no(time.Second),
			},
		},
	)
	return
}
