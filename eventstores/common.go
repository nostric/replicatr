// Package eventstores contains common things used by all of the implemented
// nostr event stores.
package eventstores

import (
	"sync"

	"realy.lol/context"
)

var (
	C      cx
	Cancel fn
	Wg     *sync.WaitGroup
)

func init() {
	Wg = &sync.WaitGroup{}
	C, Cancel = context.Cancel(context.Bg())
	return
}
