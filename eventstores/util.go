package eventstores

import (
	"bytes"

	"realy.lol/context"
	"realy.lol/lol"
)

type (
	bo = bool
	by = []byte
	st = string
	er = error
	no = int
	cx = context.T
	fn = func()
)

var (
	log, chk, errorf = lol.Main.Log, lol.Main.Check, lol.Main.Errorf
	equals           = bytes.Equal
)
