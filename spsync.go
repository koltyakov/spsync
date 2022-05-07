package spsync

import (
	"context"
	"fmt"

	"github.com/koltyakov/gosip/api"
)

type UpsertHandler func(ctx context.Context, items []ListItem) error

type DeleteHandler func(ctx context.Context, ids []int) error

// Options describes sync session and actions
type Options struct {
	SP      *api.SP
	State   *State
	EntConf *EntConf
	Upsert  UpsertHandler
	Delete  DeleteHandler
}

// Run executes sync session per entity
func Run(ctx context.Context, s *Options) (*State, error) {
	state := s.State

	// Full sync session
	if state.SyncMode == Full || state.ChangeToken == "" {
		return fullSyncSession(ctx, s)
	}

	// Incremental sync session
	if state.SyncMode == Incr && state.ChangeToken != "" {
		return incrSyncSession(ctx, s)
	}

	return state, fmt.Errorf("unknown sync mode")
}
