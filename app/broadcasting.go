package app

import (
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/eventenvelope"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/event"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/kinds"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/relayws"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/subscriptionid"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/tag"
)

// BroadcastEvent emits an event to all listeners whose filters' match, skipping all filters and actions
// it also doesn't attempt to store the event or trigger any reactions or callbacks
func (rl *Relay) BroadcastEvent(evt *event.T) {
	var remotes []string
	listeners.Range(func(ws *relayws.WebSocket, subs ListenerMap) bool {

		if ws.AuthPubKey == "" && rl.Info.Limitation.AuthRequired {
			return true
		}
		log.T.Ln("broadcasting", ws.RealRemote, ws.AuthPubKey, subs.Size())
		subs.Range(func(id string, listener *Listener) bool {
			if !listener.filters.Match(evt) {
				// log.T.F("filter doesn't match subscription %s %s\nfilters\n%s\nevent\n%s",
				// 	listener.ws.RealRemote, listener.ws.AuthPubKey,
				// 	listener.filters, evt.ToObject().String())
				return true
			}
			if kinds.IsPrivileged(evt.Kind) {
				if ws.AuthPubKey == "" {
					log.D.Ln("not broadcasting privileged event to",
						ws.RealRemote, "not authenticated")
					return true
				}
				parties := tag.T{evt.PubKey}
				pTags := evt.Tags.GetAll("p")
				for i := range pTags {
					parties = append(parties, pTags[i][1])
				}
				if !parties.Contains(ws.AuthPubKey) {
					log.D.Ln("not broadcasting privileged event to",
						ws.RealRemote, "not party to event")
					return true
				}
			}
			remotes = append(remotes, ws.RealRemote)
			chk.E(ws.WriteEnvelope(&eventenvelope.T{
				SubscriptionID: subscriptionid.T(id),
				Event:          evt},
			))
			return true
		})
		return true
	})
	if len(remotes) > 0 {
		log.T.F("sending event to subscribers %v '%s'",
			remotes, evt.ToObject().String())
	}
}
