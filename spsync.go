package spsync

import (
	"context"
	"fmt"

	"github.com/koltyakov/gosip/api"
)

type UpsertHandler func(ctx context.Context, items []AbstItem) error

type DeleteHandler func(ctx context.Context, ids []int) error

type SyncOptns struct {
	SP      *api.SP
	State   *SyncState
	EntConf *EntConf
	Upsert  UpsertHandler
	Delete  DeleteHandler
}

func Run(ctx context.Context, s *SyncOptns) (*SyncState, error) {
	state := s.State

	// Full sync session
	if state.SyncMode == FullSync || state.ChngToken == "" {
		return fullSyncSession(ctx, s)
	}

	// Incremental sync session
	if state.SyncMode == IncrSync && state.ChngToken != "" {
		return incrSyncSession(ctx, s)
	}

	return state, fmt.Errorf("unknown sync mode")
}
