package spsync

import (
	"context"
	"net/url"
	"time"

	"github.com/koltyakov/gosip/api"
)

// Full synchronization session flow
func fullSyncSession(ctx context.Context, o *Options) (*State, error) {
	var syncStart time.Time

	isBlankSync := o.State.PageToken == ""

	o.Events.FullSyncStarted(o.State.EntID, isBlankSync)

	sp := o.SP
	l := sp.Web().GetList(o.State.EntID)

	// Save current change token and timestamp for new blank sync
	if isBlankSync {
		syncStart = time.Now()
		token, err := l.Changes().GetCurrentToken()
		if err != nil {
			return o.State, err
		}
		o.State.ChangeToken = token
		o.State.SyncMode = Full
	} else {
		// For full sync continue sessions keep state values
		syncStart = o.State.SyncDate
	}

	// Sync stage dependent actions

	// Getting upsert changes
	if o.State.SyncStage == "Upsert" || o.State.SyncStage == "" {
		done := false
		for !done {
			pageToken, err := fullSyncUpsert(ctx, l, o)
			if err != nil {
				return o.State, err
			}
			o.State.PageToken = pageToken
			o.Persist(o.State)
			if pageToken == "" {
				done = true
			}
		}

		o.State.SyncStage = "Delete"
		o.Persist(o.State)
	}

	// Getting delete changes
	if o.State.SyncStage == "Delete" {
		if err := fullSyncDelete(ctx, l, o); err != nil {
			return o.State, err
		}
	}

	// Success completion state update
	o.State.PageToken = ""
	o.State.Fails = 0
	o.State.SyncDate = syncStart
	o.State.SyncStage = ""
	o.State.SyncMode = Incr

	o.Events.FullSyncFinished(o.State.EntID, isBlankSync)
	o.Persist(o.State)

	return o.State, nil
}

// Upserts processing flow
func fullSyncUpsert(ctx context.Context, l *api.List, o *Options) (string, error) {
	top := defaultPageSize
	if o.Ent.Top > 0 {
		top = o.Ent.Top
	}

	query := l.Items().Conf(api.HeadersPresets.Minimalmetadata).Top(top)
	if o.State.PageToken != "" {
		query = query.Skip(o.State.PageToken)
	}
	query = appendOData(query, o.Ent)

	o.Events.FullSyncRequest(o.State.EntID, query.ToURL())

	items, err := query.Get()
	if err != nil {
		return o.State.PageToken, err
	}

	if err := o.Upsert(ctx, itemsToUpsert(items)); err != nil {
		return o.State.PageToken, err
	}

	u, err := url.Parse(items.NextPageURL())
	if err != nil {
		return o.State.PageToken, err
	}

	token := u.Query().Get("$skiptoken")

	return token, nil
}

// Deletions processing flow
func fullSyncDelete(ctx context.Context, l *api.List, o *Options) error {
	items, err := l.Items().Conf(api.HeadersPresets.Minimalmetadata).Select("Id").Top(5000).GetAll()
	if err != nil {
		return err
	}

	// Find all missed IDs in items sequence
	var ids []int
	prevID := 0
	for _, item := range items {
		currID := item.Data().ID
		for prevID+1 != currID {
			ids = append(ids, prevID+1)
			prevID++
		}
		prevID = currID
	}

	if err := o.Delete(ctx, ids); err != nil {
		return err
	}

	return nil
}
