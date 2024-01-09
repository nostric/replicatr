package notice

import (
	"fmt"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	l "github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/labels"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/array"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/text"
)

var log = log2.GetStd()

// Envelope is a relay message intended to be shown to users in a nostr
// client interface.
type Envelope struct {
	Text string
}

func NewNoticeEnvelope(text string) (E *Envelope) {
	E = &Envelope{Text: text}
	return
}

// Label returns the label enum/type of the envelope. The relevant bytes could
// be retrieved using nip1.List[T]
func (E *Envelope) Label() string { return l.NOTICE }

func (E *Envelope) ToArray() array.T { return array.T{l.NOTICE, E.Text} }

func (E *Envelope) String() (s string) { return E.ToArray().String() }

func (E *Envelope) Bytes() (s []byte) { return E.ToArray().Bytes() }

func (E *Envelope) MarshalJSON() ([]byte, error) { return E.Bytes(), nil }

// Unmarshal the envelope.
func (E *Envelope) Unmarshal(buf *text.Buffer) (e error) {
	if E == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	// Next, find the comma after the label (note we aren't checking that only
	// whitespace intervenes because laziness, usually this is the very next
	// character).
	if e = buf.ScanUntil(','); e != nil {
		return
	}
	// Next character we find will be open quotes for the notice text.
	if e = buf.ScanThrough('"'); e != nil {
		return
	}
	var noticeText []byte
	// read the string
	if noticeText, e = buf.ReadUntil('"'); log.Fail(e) {
		return fmt.Errorf("unterminated quotes in JSON, probably truncated read")
	}
	E.Text = string(text.UnescapeByteString(noticeText))
	// log.D.F("'%s'", E.Text)
	return
}
