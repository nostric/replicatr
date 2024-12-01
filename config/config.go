package config

import (
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

type C struct {
	AppName        S             `env:"APP_NAME" default:"realy"`
	Profile        S             `env:"PROFILE" usage:"root path for all other path configurations (based on APP_NAME and OS specific location)"`
	Listen         S             `env:"LISTEN" default:"0.0.0.0" usage:"network listen address"`
	Port           N             `env:"PORT" default:"3334" usage:"port to listen on"`
	AdminUser      S             `env:"ADMIN_USER" default:"admin" usage:"admin user"`
	AdminPass      S             `env:"ADMIN_PASS" usage:"admin password"`
	LogLevel       S             `env:"LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	DbLogLevel     S             `env:"DB_LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	AuthRequired   bool          `env:"AUTH_REQUIRED" default:"false" usage:"requires auth for all access"`
	Owners         []S           `env:"OWNERS" usage:"list of npubs of users in hex format whose follow and mute list dictate accepting requests and events with AUTH_REQUIRED enabled - follows and follows follows are allowed to read/write, owners mutes events are rejected"`
	DBSizeLimit    int           `env:"DB_SIZE_LIMIT" default:"0" usage:"the number of gigabytes (1,000,000,000 bytes) we want to keep the data store from exceeding, 0 means disabled"`
	DBLowWater     int           `env:"DB_LOW_WATER" default:"60" usage:"the percentage of DBSizeLimit a GC run will reduce the used storage down to"`
	DBHighWater    int           `env:"DB_HIGH_WATER" default:"80" usage:"the trigger point at which a GC run should start if exceeded"`
	GCFrequency    int           `env:"GC_FREQUENCY" default:"3600" usage:"the frequency of checks of the current utilisation in minutes"`
	Pprof          bool          `env:"PPROF" default:"false" usage:"enable pprof on 127.0.0.1:6060"`
	MemLimit       int           `env:"MEMLIMIT" default:"250000000" usage:"set memory limit, default is 250Mb"`
	NWC            S             `env:"NWC" usage:"NWC connection string for relay to interact with an NWC enabled wallet"`
	EventStore     S             `env:"EVENTSTORE" default:"ic" usage:"type of event store ic/iconly/badger"`
	CanisterAddr   S             `env:"CANISTER_ADDR" usage:"the address of a canister"`
	CanisterId     S             `env:"CANISTER_ID" usage:"the id of a canister"`
	CanisterSecret S             `env:"SECRET_KEY" usage:"secret key for canister access"`
	PollFrequency  time.Duration `env:"POLL_FREQ" usage:"duration in 0h0m0s format between polls to canister to sync new events"`
	PollOverlap    N             `env:"POLL_OVERLAP" usage:"multiple of POLL_FREQ to back-date queries for new events to account for sync latency"`
}

func New() (cfg *C, err E) {
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
		var owners []S
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
		var val S
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
		"Environment variables that configure %s:\n\n", cfg.AppName)
	env.Usage(cfg, printer, &env.Options{SliceSep: ","})
	_, _ = fmt.Fprintf(printer,
		`CLI parameter 'help' also prints this information

.env file found at the ROOT_DIR/PROFILE path will be automatically loaded for configuration.

set these two variables for a custom load path, this file will be created on first startup.

environment overrides it and you can also edit the file to set configuration options

use the parameter 'env' to print out the current configuration to the terminal

set the environment using

	%s env > %s/%s/.env`, os.Args[0], cfg.Profile, cfg.Profile)
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

	return
}
