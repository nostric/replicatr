package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/Hubmakerlabs/replicatr/cmd/replicatrd/replicatr"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/eventstore/IC"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/eventstore/badger"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/keys"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/nip11"
	"github.com/alexflint/go-arg"
	"mleku.online/git/slog"
)

var (
	AppName = "replicatr"
	Version = "v0.0.1"
)

type ExportCmd struct {
	ToFile string `arg:"-f,--tofile" help:"write to file instead of stdout"`
}

type ImportCmd struct {
	FromFile []string `arg:"-f,--fromfile,separate" help:"read from files instead of stdin (can use flag repeatedly for multiple files)"`
}

type InitACL struct {
	Owner  string `arg:"positional,required" help:"initialize ACL configuration with an owner public key"`
	Public bool   `arg:"-p,--public" default:"false" help:"allow public read access"`
	Auth   bool   `arg:"-a,--auth" default:"false" help:"require auth for public access (recommended)"`
}

type InitCfg struct {
}

type Config struct {
	ExportCmd   *ExportCmd `json:"-" arg:"subcommand:export" help:"export database as line structured JSON"`
	ImportCmd   *ImportCmd `json:"-" arg:"subcommand:import" help:"import data from line structured JSON"`
	InitACLCmd  *InitACL   `json:"-" arg:"subcommand:initacl" help:"initialize access control configuration"`
	InitCfgCmd  *InitCfg   `json:"-" arg:"subcommand:initcfg" help:"initialize relay configuration files"`
	Listen      string     `json:"listen" arg:"-l,--listen" default:"0.0.0.0:3334" help:"network address to listen on"`
	Profile     string     `json:"-" arg:"-p,--profile" default:"replicatr" help:"profile name to use for storage"`
	Name        string     `json:"name" arg:"-n,--name" default:"replicatr relay" help:"name of relay for NIP-11"`
	Description string     `json:"description" arg:"--description" help:"description of relay for NIP-11"`
	Pubkey      string     `json:"pubkey" arg:"-k,--pubkey" help:"public key of relay operator"`
	Contact     string     `json:"contact" arg:"-c,--contact" help:"non-nostr relay operator contact details"`
	Icon        string     `json:"icon" arg:"-i,--icon" default:"https://i.nostr.build/n8vM.png" help:"icon to show on relay information pages"`
	Whitelist   []string   `arg:"-w,--whitelist,separate" help:"IP addresses that are allowed to access"`
}

var args Config

func main() {
	arg.MustParse(&args)
	var dataDirBase string
	var err error
	var log = slog.New(os.Stderr, args.Profile)
	if dataDirBase, err = os.UserHomeDir(); log.E.Chk(err) {
		os.Exit(1)
	}
	dataDir := filepath.Join(dataDirBase, args.Profile)
	log.D.F("using profile directory: '%s", args.Profile)
	var ac *replicatr.AccessControl
	if args.InitCfgCmd != nil {
		// initialize configuration with whatever has been read from the CLI.
		// include adding nip-11 configuration documents to this...
	}
	rl := replicatr.NewRelay(log, &nip11.Info{
		Name:        args.Name,
		Description: args.Description,
		PubKey:      args.Pubkey,
		Contact:     args.Contact,
		Software:    AppName,
		Version:     Version,
		Limitation: &nip11.Limits{
			MaxMessageLength: replicatr.MaxMessageSize,
			Oldest:           1640305963,
		},
		RelayCountries: nil,
		LanguageTags:   nil,
		Tags:           nil,
		PostingPolicy:  "",
		PaymentsURL:    "",
		Fees:           &nip11.Fees{},
		Icon:           args.Icon,
	}, args.Whitelist, ac)
	aclPath := filepath.Join(dataDir, replicatr.ACLfilename)
	// initialise ACL if command is called. Note this will overwrite an existing
	// configuration.
	if args.InitACLCmd != nil {
		if !keys.IsValid32ByteHex(args.InitACLCmd.Owner) {
			log.E.Ln("invalid owner public key")
			os.Exit(1)
		}
		rl.AccessControl = &replicatr.AccessControl{
			Users: []*replicatr.UserID{
				{
					Role:      replicatr.RoleOwner,
					PubKeyHex: args.InitACLCmd.Owner,
				},
			},
			Public:     args.InitACLCmd.Public,
			PublicAuth: args.InitACLCmd.Auth,
		}
		rl.Info.Limitation.AuthRequired = args.InitACLCmd.Auth
		// if the public flag is set, add an empty reader to signal public reader
		if err = rl.SaveACL(aclPath); rl.E.Chk(err) {
			panic(err)
			// this is probably a fatal error really
		}
		log.I.Ln("access control base configuration saved and ready to use")
	}
	// load access control list
	if err = rl.LoadACL(aclPath); rl.W.Chk(err) {
		rl.W.Ln("no access control configured for relay")
	}
	rl.Info.AddNIPs(
		nip11.BasicProtocol.Number,            // events, envelopes and filters
		nip11.FollowList.Number,               // follow lists
		nip11.EncryptedDirectMessage.Number,   // encrypted DM
		nip11.MappingNostrKeysToDNS.Number,    // DNS
		nip11.EventDeletion.Number,            // event delete
		nip11.RelayInformationDocument.Number, // relay information document
		nip11.NostrMarketplace.Number,         // marketplace
		nip11.Reposts.Number,                  // reposts
		nip11.Bech32EncodedEntities.Number,    // bech32 encodings
		nip11.LongFormContent.Number,          // long form
		nip11.PublicChat.Number,               // public chat
		nip11.UserStatuses.Number,             // user statuses
		nip11.Authentication.Number,           // auth
		nip11.CountingResults.Number,          // count requests
	)
	db := &IC.Backend{
		Badger: &badger.Backend{
			Path: dataDir,
			Log:  slog.New(os.Stderr, "replicatr-badger"),
		},
	}
	if err = db.Init(); rl.E.Chk(err) {
		rl.E.F("unable to start database: '%s'", err)
		os.Exit(1)
	}
	rl.StoreEvent = append(rl.StoreEvent, db.SaveEvent)
	rl.QueryEvents = append(rl.QueryEvents, db.QueryEvents)
	rl.CountEvents = append(rl.CountEvents, db.CountEvents)
	rl.DeleteEvent = append(rl.DeleteEvent, db.DeleteEvent)
	rl.RejectFilter = append(rl.RejectFilter, rl.FilterAccessControl)
	rl.RejectCountFilter = append(rl.RejectCountFilter, rl.FilterAccessControl)
	switch {
	case args.ImportCmd != nil:
		rl.Import(db.Badger, args.ImportCmd.FromFile)
	case args.ExportCmd != nil:
		rl.Export(db.Badger, args.ExportCmd.ToFile)
	default:
		rl.I.Ln("listening on", args.Listen)
		rl.E.Chk(http.ListenAndServe(args.Listen, rl))
	}
}
