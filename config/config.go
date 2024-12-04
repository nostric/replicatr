package config

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"go-simpler.org/env"

	"realy.lol/appdata"
	"realy.lol/apputil"
	"realy.lol/config"
	"realy.lol/sha256"
)

//go:embed version
var Version string

// C is the configuration items for replicatr.
type C struct {
	AppName        st            `env:"APP_NAME" default:"replicatr" json:"app_name,omitempty"`
	Description    st            `env:"APP_DESCRIPTION" default:"a nostr relay that uses an Internet Computer canister as a shared event store" json:"description,omitempty"`
	Profile        st            `env:"PROFILE" usage:"root path for all other path configurations (based on APP_NAME and OS specific location)" json:"profile,omitempty"`
	Listen         st            `env:"LISTEN" default:"0.0.0.0" usage:"network listen address" json:"listen,omitempty"`
	Port           no            `env:"PORT" default:"3334" usage:"port to listen on" json:"port,omitempty"`
	AdminUser      st            `env:"ADMIN_USER" default:"admin" usage:"admin user" json:"admin_user,omitempty"`
	AdminPass      st            `env:"ADMIN_PASS" usage:"admin password" json:"admin_pass,omitempty"`
	LogLevel       st            `env:"LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace" json:"log_level,omitempty"`
	DbLogLevel     st            `env:"DB_LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace" json:"db_log_level,omitempty"`
	AuthRequired   bool          `env:"AUTH_REQUIRED" default:"false" usage:"requires auth for all access" json:"auth_required,omitempty"`
	Owners         []st          `env:"OWNERS" usage:"list of npubs of users in hex format whose follow and mute list dictate accepting requests and events with AUTH_REQUIRED enabled - follows and follows follows are allowed to read/write, owners mutes events are rejected" json:"owners,omitempty"`
	DBSizeLimit    int           `env:"DB_SIZE_LIMIT" default:"0" usage:"the number of gigabytes (1,000,000,000 bytes) we want to keep the data store from exceeding, 0 means disabled" json:"db_size_limit,omitempty"`
	DBLowWater     int           `env:"DB_LOW_WATER" default:"60" usage:"the percentage of DBSizeLimit a GC run will reduce the used storage down to" json:"db_low_water,omitempty"`
	DBHighWater    int           `env:"DB_HIGH_WATER" default:"80" usage:"the trigger point at which a GC run should start if exceeded" json:"db_high_water,omitempty"`
	GCFrequency    int           `env:"GC_FREQUENCY" default:"3600" usage:"the frequency of checks of the current utilisation in minutes" json:"gc_frequency,omitempty"`
	Pprof          bool          `env:"PPROF" default:"false" usage:"enable pprof on 127.0.0.1:6060" json:"pprof,omitempty"`
	MemLimit       int           `env:"MEMLIMIT" default:"250000000" usage:"set memory limit, default is 250Mb" json:"mem_limit,omitempty"`
	NWC            st            `env:"NWC" usage:"NWC connection string for relay to interact with an NWC enabled wallet" json:"nwc,omitempty"`
	EventStore     st            `env:"EVENTSTORE" default:"ic" usage:"type of event store ic/iconly/badger" json:"event_store,omitempty"`
	CanisterAddr   st            `env:"CANISTER_ADDR" usage:"the address of a canister" json:"canister_addr,omitempty"`
	CanisterId     st            `env:"CANISTER_ID" usage:"the id of a canister" json:"canister_id,omitempty"`
	CanisterSecret st            `env:"SECRET_KEY" usage:"secret key for canister access" json:"canister_secret,omitempty"`
	PollFrequency  time.Duration `env:"POLL_FREQ" default:"5s" usage:"duration in 0h0m0s format between polls to canister to sync new events" json:"poll_frequency,omitempty"`
	PollOverlap    no            `env:"POLL_OVERLAP" usage:"multiple of POLL_FREQ to back-date queries for new events to account for sync latency" json:"poll_overlap,omitempty"`
}

func New() (cfg *C, err er) {
	cfg = &C{}
	if err = env.Load(cfg, nil); chk.T(err) {
		return
	}
	if cfg.Profile == "" {
		cfg.Profile = appdata.Dir(cfg.AppName, true)
	}
	envPath := filepath.Join(cfg.Profile, ".env")
	if apputil.FileExists(envPath) {
		var e config.Env
		if e, err = config.GetEnv(envPath); chk.T(err) {
			return
		}
		if err = env.Load(cfg, &env.Options{Source: e}); chk.E(err) {
			return
		}
		var owners []st
		// remove empties if any
		for _, o := range cfg.Owners {
			if len(o) == sha256.Size*2 {
				owners = append(owners, o)
			}
		}
		cfg.Owners = owners
	}
	return
}

// HelpRequested returns true if any of the common types of help invocation are
// found as the first command line parameter/flag.
func HelpRequested() (help bool) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "help", "-h", "--h", "-help", "--help", "?":
			help = true
		}
	}
	return
}

func GetEnv() (requested bool) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "env":
			requested = true
		}
	}
	return
}

func PrintEnv(cfg *C, printer io.Writer) {
	t := reflect.TypeOf(*cfg)

	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Tag.Get("env")
		v := reflect.ValueOf(*cfg).Field(i).Interface()
		var val st
		switch v.(type) {
		case string:
			val = v.(string)
		case int, bool:
			val = fmt.Sprint(v)
		case []string:
			arr := v.([]string)
			if len(arr) > 0 {
				val = strings.Join(arr, ",")
			}
		}
		fmt.Fprintf(printer, "%s=%v\n", k, val)
	}
}

// PrintHelp outputs a help text listing the configuration options and default
// values to a provided io.Writer (usually os.Stderr or os.Stdout).
func PrintHelp(cfg *C, printer io.Writer) {
	_, _ = fmt.Fprintf(printer,
		`%s -- %s

	CLI parameter 'help' also prints this information

	.env file found at the ROOT_DIR/PROFILE path will be automatically loaded for configuration.

	set these two variables for a custom load path, this file will be created on first startup.

	environment overrides it and you can also edit the file to set configuration options

	use the parameter 'env' to print out the current configuration to the terminal

	set the environment using

		%s env > %s/%s/.env

`, cfg.AppName, cfg.Description, cfg.Profile, cfg.Profile, Version)
	_, _ = fmt.Fprintf(printer, `other commands you can invoke from command line args:

	%s pubkey
		prints the public key being used by the relay for authentication, required for the canister
		owner to authorize using "addrelay" or remove with "removerelay"
	
	%s addrelay <pubkey> <admin: true/false>
		adds a relay pubkey to the list of relay public keys permitted to use the canister
	
	%s removerelay <pubkey>
		removes a relay from the list of relay public keys permitted to use the canister
	
	%s getpermission
		reports the access level of this relay to the canister

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	_, _ = fmt.Fprintf(printer,
		"Environment variables that configure %s:\n\n", cfg.AppName)
	env.Usage(cfg, printer, &env.Options{SliceSep: ","})
	_, _ = fmt.Fprintln(printer)
	return
}
