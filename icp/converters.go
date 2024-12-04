package icp

import (
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"

	"github.com/Hubmakerlabs/replicatr/canister/candid"
)

// TagMapToKV converts tags, in the format used with realy the first field is the key and the
// remainder are values, the result is a candid.KeyPair with the first element of each tag is
// the candid.KeyValuePair Key, and the remainder of each tag is the Value.
//
// Used for the candid.Filter.
func TagMapToKV(t *tags.T) (keys []candid.KeyValuePair) {
	keys = make([]candid.KeyValuePair, 0, t.Len())
	for _, tt := range t.F() {
		k := st(tt.F()[0])
		v := tt.F()[1:]
		var vals []st
		for _, vv := range v {
			vals = append(vals, st(vv))
		}
		keys = append(keys, candid.KeyValuePair{Key: k, Value: vals})
	}
	return
}

// FilterToCandid converts a nostr filter.T to a candid.Filter.
//
// There is no reverse conversion because it is unnecessary.
func FilterToCandid(f *filter.T) (result *candid.Filter) {
	result = &candid.Filter{
		IDs:     f.IDs.ToStringSlice(),
		Kinds:   f.Kinds.ToUint16(),
		Authors: f.Authors.ToStringSlice(),
		Tags:    TagMapToKV(f.Tags),
		Search:  st(f.Search),
	}
	if f.Since != nil {
		result.Since = int64(f.Since.I64())
	} else {
		result.Since = -1
	}

	if f.Until != nil {
		result.Until = f.Until.I64()
	} else {
		result.Until = -1
	}

	if f.Limit != nil {
		result.Limit = int64(*f.Limit)
	} else {
		result.Limit = 500
	}

	return
}

// EventToCandid converts an nostr event.T to a candid.Event.
func EventToCandid(e *event.T) candid.Event {
	return candid.Event{
		ID:        hex.Enc(e.ID),
		Pubkey:    hex.Enc(e.PubKey),
		CreatedAt: e.CreatedAt.I64(),
		Kind:      e.Kind.ToU16(),
		Tags:      e.Tags.ToStringSlice(),
		Content:   st(e.Content),
		Sig:       hex.Enc(e.Sig),
	}
}

// CandidToEvent converts a candid.Event to a nostr event.T.
func CandidToEvent(e *candid.Event) *event.T {
	tt := tags.NewWithCap(len(e.Tags))
	for i := range e.Tags {
		t := tag.New(e.Tags[i]...)
		tt.AppendTags(t)
	}
	var err er
	var id, pk, sig by
	if id, err = hex.Dec(e.ID); chk.E(err) {
		return nil
	}
	if pk, err = hex.Dec(e.Pubkey); chk.E(err) {
		return nil
	}
	if sig, err = hex.Dec(e.Sig); chk.E(err) {
		return nil
	}
	return &event.T{
		ID:        id,
		PubKey:    pk,
		CreatedAt: timestamp.FromUnix(e.CreatedAt),
		Kind:      kind.New(e.Kind),
		Tags:      tt,
		Content:   by(e.Content),
		Sig:       sig,
	}
}
