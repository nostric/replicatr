package replicatr

import (
	"context"
	"errors"
	"sync"

	"github.com/Hubmakerlabs/replicatr/pkg/nostr/normalize"
	"github.com/nbd-wtf/go-nostr"
)

func (rl *Relay) handleRequest(ctx context.Context, id string,
	eose *sync.WaitGroup, ws *WebSocket, f *nostr.Filter) (e error) {

	defer eose.Done()
	// overwrite the filter (for example, to eliminate some kinds or that we
	// know we don't support)
	for _, ovw := range rl.OverwriteFilter {
		ovw(ctx, f)
	}
	if f.Limit < 0 {
		e = errors.New("blocked: filter invalidated")
		rl.E.Chk(e)
		return
	}
	// then check if we'll reject this filter (we apply this after overwriting
	// because we may, for example, remove some things from the incoming filters
	// that we know we don't support, and then if the end result is an empty
	// filter we can just reject it)
	for _, reject := range rl.RejectFilter {
		if rej, msg := reject(ctx, f); rej {
			rl.E.Chk(ws.WriteJSON(nostr.NoticeEnvelope(msg)))
			return errors.New(normalize.OKMessage(msg, "blocked"))
		}
	}
	// run the functions to query events (generally just one,
	// but we might be fetching stuff from multiple places)
	eose.Add(len(rl.QueryEvents))
	for _, query := range rl.QueryEvents {
		var ch chan *nostr.Event
		if ch, e = query(ctx, f); rl.E.Chk(e) {
			rl.E.Chk(ws.WriteJSON(nostr.NoticeEnvelope(e.Error())))
			eose.Done()
			continue
		}
		go func(ch chan *nostr.Event) {
			for event := range ch {
				for _, ovw := range rl.OverwriteResponseEvent {
					ovw(ctx, event)
				}
				rl.E.Chk(ws.WriteJSON(nostr.EventEnvelope{
					SubscriptionID: &id,
					Event:          *event,
				}))
			}
			eose.Done()
		}(ch)
	}
	return nil
}

func (rl *Relay) handleCountRequest(ctx context.Context, ws *WebSocket,
	filter *nostr.Filter) (subtotal int64) {
	// overwrite the filter (for example, to eliminate some kinds or tags that
	// we know we don't support)
	for _, ovw := range rl.OverwriteCountFilter {
		ovw(ctx, filter)
	}
	// then check if we'll reject this filter
	for _, reject := range rl.RejectCountFilter {
		if rej, msg := reject(ctx, filter); rej {
			rl.E.Chk(ws.WriteJSON(nostr.NoticeEnvelope(msg)))
			return 0
		}
	}
	// run the functions to count (generally it will be just one)
	var e error
	var res int64
	for _, count := range rl.CountEvents {
		if res, e = count(ctx, filter); rl.E.Chk(e) {
			rl.E.Chk(ws.WriteJSON(nostr.NoticeEnvelope(e.Error())))
		}
		subtotal += res
	}
	return
}
