package spsync

import (
	"context"
	"fmt"

	"github.com/koltyakov/gosip/api"
)

type UpsertHandler func(ctx context.Context, items []Item) error

type DeleteHandler func(ctx context.Context, ids []int) error

// Options describes sync session and actions
type Options struct {
	SP      *api.SP
	State   *State
	Ent     *Ent
	Upsert  UpsertHandler
	Delete  DeleteHandler
	Events  *Events
	Persist func(s *State) error
}

// Run executes sync session per entity
func Run(ctx context.Context, o *Options) (*State, error) {
	opts, err := ensureOptions(o)
	if err != nil {
		return o.State, err
	}

	// Full sync session
	if opts.State.SyncMode == Full || opts.State.ChangeToken == "" {
		return fullSyncSession(ctx, opts)
	}

	// Incremental sync session
	if opts.State.SyncMode == Incr && opts.State.ChangeToken != "" {
		return incrSyncSession(ctx, opts)
	}

	return opts.State, fmt.Errorf("can't resolve sync strategy")
}

// Validates and extends options with defaults
func ensureOptions(o *Options) (*Options, error) {
	if o.SP == nil {
		return o, fmt.Errorf("SP is not provided")
	}
	if o.State == nil {
		return o, fmt.Errorf("state is not provided")
	}
	if o.Ent == nil {
		return o, fmt.Errorf("entity configuration is not provided")
	}
	if o.Upsert == nil {
		return o, fmt.Errorf("upsert handler is not provided")
	}
	if o.Delete == nil {
		return o, fmt.Errorf("delete handler is not provided")
	}
	if o.State.EntID == "" {
		return o, fmt.Errorf("entity ID is not provided in state")
	}

	if o.Persist == nil {
		// Persisting state incrementally is recommended
		// so paged content state is persisted in case of error
		// and next start continues from last saved step
		o.Persist = func(s *State) error {
			fmt.Println("WARN: persist handler is not provided, state won't be saved incrementally")
			return nil
		}
	}

	// Populates empty events if not provided
	o.Events = ensureEvents(o.Events)

	return o, nil
}
