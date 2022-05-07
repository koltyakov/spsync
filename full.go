package spsync

import (
	"context"
	"net/url"
	"time"

	"github.com/koltyakov/gosip/api"
)

// Full synchronization session flow
func fullSyncSession(ctx context.Context, s *Options) (*State, error) {
	var syncDate time.Time
	var chngToken string

	state := s.State

	isBlankSync := state.PageToken == ""

	sp := s.SP
	ent := sp.Web().GetList(s.State.EntID)

	// Save current change token and timestamp for new blank sync
	if isBlankSync {
		syncDate = time.Now()
		token, err := ent.Changes().GetCurrentToken()
		if err != nil {
			return state, err
		}
		chngToken = token
		state.SyncMode = Full
	} else {
		// For full sync continue sessions keep state values
		syncDate = state.SyncDate
		chngToken = state.ChangeToken
	}

	// Sync stage dependent actions

	// Getting upsert changes
	if state.SyncStage == "Upsert" || state.SyncStage == "" {
		completed := false
		for !completed {
			pageToken, err := fullSyncUpsert(ctx, ent, state.PageToken, s.EntConf, s.Upsert)
			if err != nil {
				return state, err
			}
			state.PageToken = pageToken
			if pageToken == "" {
				completed = true
			}
		}

		state.SyncStage = "Delete"
	}

	// Getting delete changes
	if state.SyncStage == "Delete" {
		if err := fullSyncDelete(ctx, ent, s.EntConf, s.Delete); err != nil {
			return state, err
		}
	}

	// Success completion state update
	state.PageToken = ""
	state.Fails = 0
	state.SyncDate = syncDate
	state.ChangeToken = chngToken
	state.SyncStage = ""
	state.SyncMode = Incr

	return state, nil
}

// Upserts processing flow
func fullSyncUpsert(ctx context.Context, e *api.List, token string, c *EntConf, up UpsertHandler) (string, error) {
	top := deafultPageSize
	if c.Top > 0 {
		top = c.Top
	}

	query := e.Items().Conf(api.HeadersPresets.Minimalmetadata).Top(top)
	if token != "" {
		query = query.Skip(token)
	}
	query = appendOData(query, c)

	items, err := query.Get()
	if err != nil {
		return token, err
	}

	if err := up(ctx, itemsToUpsert(items)); err != nil {
		return token, err
	}

	u, err := url.Parse(items.NextPageURL())
	if err != nil {
		return token, err
	}

	token = u.Query().Get("$skiptoken")

	return token, nil
}

// Deletions processing flow
func fullSyncDelete(ctx context.Context, e *api.List, c *EntConf, del DeleteHandler) error {
	items, err := e.Items().Conf(api.HeadersPresets.Minimalmetadata).Select("Id").Top(5000).GetAll()
	if err != nil {
		return err
	}

	var ids []int
	prevID := 0
	for _, item := range items {
		currID := item.Data().ID
		if prevID+1 != currID {
			ids = append(ids, currID)
		}
		prevID = currID
	}

	if err := del(ctx, ids); err != nil {
		return err
	}

	return nil
}
