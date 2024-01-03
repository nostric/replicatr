package replicatr

import (
	"context"
	"net/http"
	"os"
	"time"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	
	"github.com/fasthttp/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/puzpuzpuz/xsync/v2"
)

const (
	WriteWait             = 10 * time.Second
	PongWait              = 60 * time.Second
	PingPeriod            = 30 * time.Second
	ReadBufferSize        = 4096
	WriteBufferSize       = 4096
	MaxMessageSize  int64 = 512000 // ???
)

func NewRelay(appName string) (r *Relay) {
	r = &Relay{
		Log: log2.New(os.Stderr, appName, 0),
		Info: &nip11.RelayInformationDocument{
			Software:      "https://github.com/Hubmakerlabs/replicatr/cmd/khatru",
			Version:       "n/a",
			SupportedNIPs: make([]int, 0),
		},
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		clients:        xsync.NewTypedMapOf[*websocket.Conn, struct{}](pointerHasher[websocket.Conn]),
		serveMux:       &http.ServeMux{},
		WriteWait:      WriteWait,
		PongWait:       PongWait,
		PingPeriod:     PingPeriod,
		MaxMessageSize: MaxMessageSize,
	}
	return
}

type (
	RejectEvent  func(ctx context.Context, event *nostr.Event) (reject bool, msg string)
	RejectFilter func(ctx context.Context, filter *nostr.Filter) (reject bool, msg string)
)

type Relay struct {
	ServiceURL                string
	RejectEvent               []RejectEvent
	RejectFilter              []RejectFilter
	RejectCountFilter         []func(ctx context.Context, filter *nostr.Filter) (reject bool, msg string)
	OverwriteDeletionOutcome  []func(ctx context.Context, target *nostr.Event, deletion *nostr.Event) (acceptDeletion bool, msg string)
	OverwriteResponseEvent    []func(ctx context.Context, event *nostr.Event)
	OverwriteFilter           []func(ctx context.Context, filter *nostr.Filter)
	OverwriteCountFilter      []func(ctx context.Context, filter *nostr.Filter)
	OverwriteRelayInformation []func(ctx context.Context, r *http.Request, info nip11.RelayInformationDocument) nip11.RelayInformationDocument
	StoreEvent                []func(ctx context.Context, event *nostr.Event) error
	DeleteEvent               []func(ctx context.Context, event *nostr.Event) error
	QueryEvents               []func(ctx context.Context, filter *nostr.Filter) (chan *nostr.Event, error)
	CountEvents               []func(ctx context.Context, filter *nostr.Filter) (int64, error)
	OnConnect                 []func(ctx context.Context)
	OnDisconnect              []func(ctx context.Context)
	OnEventSaved              []func(ctx context.Context, event *nostr.Event)
	// editing info will affect
	Info *nip11.RelayInformationDocument
	Log  *log2.Logger
	// for establishing websockets
	upgrader websocket.Upgrader
	// keep a connection reference to all connected clients for Server.Shutdown
	clients *xsync.MapOf[*websocket.Conn, struct{}]
	// in case you call Server.Start
	Addr       string
	serveMux   *http.ServeMux
	httpServer *http.Server
	// websocket options
	WriteWait      time.Duration // Time allowed to write a message to the peer.
	PongWait       time.Duration // Time allowed to read the next pong message from the peer.
	PingPeriod     time.Duration // Send pings to peer with this period. Must be less than pongWait.
	MaxMessageSize int64         // Maximum message size allowed from peer.
}
