package enveloper

import (
	"encoding/json"
	"fmt"

	"github.com/Hubmakerlabs/replicatr/pkg/go-nostr/envelopes"
	"github.com/Hubmakerlabs/replicatr/pkg/interfaces/arrayer"
	"github.com/Hubmakerlabs/replicatr/pkg/interfaces/buffer"
	"github.com/Hubmakerlabs/replicatr/pkg/interfaces/byter"
	"github.com/Hubmakerlabs/replicatr/pkg/labeler"
)

// I interface for envelopes
//
// Note that the Unmarshal function is not UnmarshalJSON for a specific reason -
// it is impossible to implement a typed JSON unmarshaler in Go for an array
// type because it must by definition have a sentinel field which in the case of
// nostr is the Label. Objects have a defined collection of recognised labels
// and with omitempty marking the mandatory ones, acting as a "kind" of
// sentinel.
type I interface {
	envelopes.E // transitional
	labeler.I
	fmt.Stringer
	byter.I
	json.Marshaler
	buffer.Unmarshaler
	arrayer.I
}