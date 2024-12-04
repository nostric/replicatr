// Package candid is a library defining data structures used with the Internet Computer Protocol
// nostr event store canister.
package candid

type (
	// Event is the go representation of a candid encoding of a nostr event.
	Event struct {
		ID        string     `ic:"id"`
		Pubkey    string     `ic:"pubkey"`
		CreatedAt int64      `ic:"created_at"`
		Kind      uint16     `ic:"kind"`
		Tags      [][]string `ic:"tags"`
		Content   string     `ic:"content"`
		Sig       string     `ic:"sig"`
	}

	// Filter is the go representation of a candid encoding of a nostr filter.
	Filter struct {
		IDs     []string       `ic:"ids"`
		Kinds   []uint16       `ic:"kinds"`
		Authors []string       `ic:"authors"`
		Tags    []KeyValuePair `ic:"tags"`
		Since   int64          `ic:"since"`
		Until   int64          `ic:"until"`
		Limit   int64          `ic:"limit"`
		Search  string         `ic:"search"`
	}

	// KeyValuePair is a simple struct representing a nostr filter tag item.
	KeyValuePair struct {
		Key   string   `ic:"key"`
		Value []string `ic:"value"`
	}
)
