package agent

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	agent_go "github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	sec "github.com/aviate-labs/secp256k1"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

type (
	Event struct {
		ID        string     `ic:"id"`
		Pubkey    string     `ic:"pubkey"`
		CreatedAt int64      `ic:"created_at"`
		Kind      uint16     `ic:"kind"`
		Tags      [][]string `ic:"tags"`
		Content   string     `ic:"content"`
		Sig       string     `ic:"sig"`
	}
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
	KeyValuePair struct {
		Key   string   `ic:"key"`
		Value []string `ic:"value"`
	}
)

type Backend struct {
	Ctx        context.T
	Agent      *agent_go.Agent
	CanisterID principal.Principal
}

func (b *Backend) GetCandidEvent(filter *Filter) ([]Event, error) {
	methodName := "get_events"
	args := []any{*filter, time.Now().UnixNano()}
	var result []Event
	err := b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return nil, err
	}
	return result, err
}

func (b *Backend) CountCandidEvent(filter *Filter) (int, error) {
	methodName := "count_events"
	args := []any{*filter, time.Now().UnixNano()}
	var result int
	err := b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return -1, err
	}
	return result, err
}

func (b *Backend) ClearCandidEvents() (err error) {
	methodName := "clear_events"
	var result *string
	args := []any{time.Now().UnixNano()}
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Clear Events")
	}
	return
}

func (b *Backend) SaveCandidEvent(event Event) (err error) {
	methodName := "save_event"
	var result *string
	args := []any{event, time.Now().UnixNano()}
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Store Event")
	}
	return
}

func (b *Backend) DeleteCandidEvent(event Event) (err error) {
	methodName := "delete_event"
	args := []any{event, time.Now().UnixNano()}
	var result *string
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Delete Event")
	}
	return
}

func (b *Backend) RemoveUser(pubKey string) (err error) {
	methodName := "remove_user"
	args := []any{pubKey, time.Now().UnixNano()}
	var result *string
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return
	} else if result != nil {
		err = log.E.Err("failed to remove user")
		return
	}
	return nil
}

func (b *Backend) GetPermission() (result string, err error) {
	methodName := "get_permission"
	args := []any{time.Now().UnixNano()}
	err = b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return "", err
	}
	return result, nil
}

func (b *Backend) AddUser(pubKey string, perm bool) (err error) {
	methodName := "add_user"
	args := []any{pubKey, perm, time.Now().UnixNano()}
	var result *string
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return
	} else if result != nil {
		err = log.E.Err("failed to add user")
		return
	}
	return nil
}

func (b *Backend) QueryEvents(f *filter.T) (ch event.C, err error) {
	ch = make(event.C)
	go func() {
		if f == nil {
			err = log.E.Err("nil filter for query")
			return
		}
		var candidEvents []Event
		if candidEvents, err = b.GetCandidEvent(FilterToCandid(f)); err != nil {
			split := strings.Split(err.Error(), "Error: ")
			if len(split) == 2 && split[1] != "No events found" {
				log.E.F("IC error: %s", split[1])
			}
			return
		}
		log.I.Ln("got", len(candidEvents), "events")
		for i, e := range candidEvents {
			select {
			case <-b.Ctx.Done():
				return
			default:
			}
			log.T.Ln("sending event", i)
			ch <- CandidToEvent(&e)
		}
		log.T.Ln("done sending events")
	}()
	return
}

func (b *Backend) SaveEvent(e *event.T) (err error) {
	select {
	case <-b.Ctx.Done():
		return
	default:
	}
	if err = b.SaveCandidEvent(EventToCandid(e)); chk.E(err) {
		return
	}
	return
}

func (b *Backend) DeleteEvent(ev *event.T) (err error) {
	select {
	case <-b.Ctx.Done():
		return
	default:
	}
	if err = b.DeleteCandidEvent(EventToCandid(ev)); chk.E(err) {
		return
	}
	return
}

func (b *Backend) CountEvents(f *filter.T) (count int, err error) {
	if f == nil {
		err = log.E.Err("nil filter for count query")
		return
	}
	count, err = b.CountCandidEvent(FilterToCandid(f))
	return
}
func (b *Backend) ClearEvents() (err error) {
	err = b.ClearCandidEvents()
	return
}

func New(c context.T, cid, canAddr, secKey string) (a *Backend, err error) {
	log.D.Ln("setting up IC backend to", canAddr, cid)
	a = &Backend{Ctx: c}
	localReplicaURL, _ := url.Parse(canAddr)
	secKeyBytes, err := hex.Dec(secKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key hex: %s", err)
	}
	privKey, _ := sec.PrivKeyFromBytes(sec.S256(), secKeyBytes)
	id, _ := identity.NewSecp256k1Identity(privKey)
	cfg := agent_go.Config{FetchRootKey: true,
		ClientConfig: &agent_go.ClientConfig{Host: localReplicaURL}, Identity: id}
	if a.Agent, err = agent_go.New(cfg); chk.E(err) {
		return
	}
	if a.CanisterID, err = principal.Decode(cid); chk.E(err) {
		return
	}
	log.D.Ln("successfully connected to IC backend")
	return
}

func TagMapToKV(t *tags.T) (keys []KeyValuePair) {
	keys = make([]KeyValuePair, 0, t.Len())
	for _, tt := range t.F() {
		k := S(tt.F()[0])
		v := tt.F()[1:]
		var vals []S
		for _, vv := range v {
			vals = append(vals, S(vv))
		}
		keys = append(keys, KeyValuePair{Key: k, Value: vals})
	}
	return
}

func FilterToCandid(f *filter.T) (result *Filter) {
	result = &Filter{
		IDs:     f.IDs.ToStringSlice(),
		Kinds:   f.Kinds.ToUint16(),
		Authors: f.Authors.ToStringSlice(),
		Tags:    TagMapToKV(f.Tags),
		Search:  S(f.Search),
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

func EventToCandid(e *event.T) Event {
	return Event{
		ID:        hex.Enc(e.ID),
		Pubkey:    hex.Enc(e.PubKey),
		CreatedAt: e.CreatedAt.I64(),
		Kind:      e.Kind.ToU16(),
		Tags:      e.Tags.ToStringSlice(),
		Content:   S(e.Content),
		Sig:       hex.Enc(e.Sig),
	}
}

func CandidToEvent(e *Event) *event.T {
	tt := tags.NewWithCap(len(e.Tags))
	for i := range e.Tags {
		t := tag.New(e.Tags[i]...)
		tt.AppendTags(t)
	}
	var err E
	var id, pk, sig B
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
		Content:   B(e.Content),
		Sig:       sig,
	}
}
