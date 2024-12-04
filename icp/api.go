package icp

import (
	"fmt"
	"net/url"
	"time"

	agent_go "github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	sec "github.com/aviate-labs/secp256k1"
	"realy.lol/hex"

	"github.com/Hubmakerlabs/replicatr/canister/candid"
)

type T struct {
	Agent      *agent_go.Agent
	CanisterID principal.Principal
}

func New(cid, canAddr, secKey string) (a *T, err er) {
	log.D.Ln("setting up IC backend to", canAddr, cid)
	a = &T{}
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

func (b *T) GetCandidEvent(filter *candid.Filter) ([]candid.Event, er) {
	methodName := "get_events"
	args := []any{*filter, time.Now().UnixNano()}
	var result []candid.Event
	err := b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return nil, err
	}
	return result, err
}

func (b *T) CountCandidEvent(filter *candid.Filter) (int, er) {
	methodName := "count_events"
	args := []any{*filter, time.Now().UnixNano()}
	var result int
	err := b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return -1, err
	}
	return result, err
}

func (b *T) ClearCandidEvents() (err er) {
	methodName := "clear_events"
	var result *string
	args := []any{time.Now().UnixNano()}
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Clear Events")
	}
	return
}

func (b *T) SaveCandidEvent(event candid.Event) (err er) {
	methodName := "save_event"
	var result *string
	args := []any{event, time.Now().UnixNano()}
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Store Event")
	}
	return
}

func (b *T) DeleteCandidEvent(event st) (err er) {
	methodName := "delete_event"
	args := []any{event, time.Now().UnixNano()}
	var result *string
	err = b.Agent.Call(b.CanisterID, methodName, args, []any{&result})
	if err == nil && result != nil {
		err = log.E.Err("Unable to Delete Event")
	}
	return
}

func (b *T) RemoveUser(pubKey string) (err er) {
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

func (b *T) GetPermission() (result string, err er) {
	methodName := "get_permission"
	args := []any{time.Now().UnixNano()}
	err = b.Agent.Query(b.CanisterID, methodName, args, []any{&result})
	if err != nil {
		return "", err
	}
	return result, nil
}

func (b *T) AddUser(pubKey string, perm bool) (err er) {
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

func (b *T) ClearEvents() (err er) {
	err = b.ClearCandidEvents()
	return
}
