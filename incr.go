package spsync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/koltyakov/gosip/api"
)

// Incremental synchronization session flow
func incrSyncSession(ctx context.Context, o *Options) (*State, error) {
	syncDate := time.Now()

	o.Events.IncrSyncStarted(o.State.EntID)

	sp := o.SP
	ent := sp.Web().GetList(o.State.EntID)

	tillToken, err := ent.Changes().GetCurrentToken()
	if err != nil {
		return o.State, err
	}

	completed := false
	for !completed {
		changeToken, changes, err := incrSyncPaged(ctx, ent, tillToken, o)
		if err != nil {
			return o.State, err
		}
		o.State.ChangeToken = changeToken
		o.Persist(o.State)
		if changes == 0 {
			completed = true
		}
	}

	// Success completion state update
	o.State.PageToken = ""
	o.State.Fails = 0
	o.State.SyncDate = syncDate
	o.State.ChangeToken = tillToken
	o.State.SyncStage = ""

	o.Events.IncrSyncFinished(o.State.EntID)
	o.Persist(o.State)

	return o.State, nil
}

// Change API paged responce processing
func incrSyncPaged(ctx context.Context, l *api.List, endToken string, o *Options) (string, int, error) {
	o.Events.IncrSyncRequest(o.State.EntID, o.State.ChangeToken, endToken)

	// Default 100 items per page is used for getting changes
	// the page size increase is not recommended as change API returns only IDs
	// and when IDs are used to construct requests to get specific items
	changes, _ := l.Changes().Top(100).GetChanges(&api.ChangeQuery{
		ChangeTokenStart: o.State.ChangeToken,
		ChangeTokenEnd:   endToken,
		Item:             true,
		Restore:          true,
		Add:              true,
		DeleteObject:     true,
		Update:           true,
		SystemUpdate:     true,
	})

	changesCnt := len(changes.Data())

	if changesCnt == 0 {
		return "", 0, nil
	}

	deleteChangeType := 3

	// Upserted
	var upsertIds []int
	for _, ch := range changes.Data() {
		if ch.ChangeType != deleteChangeType {
			upsertIds = append(upsertIds, ch.ItemID)
		}
	}

	if len(upsertIds) > 0 {
		query := l.Items().Conf(api.HeadersPresets.Minimalmetadata).Top(len(upsertIds))
		query = appendOData(query, o.Ent)

		var filters []string
		for _, id := range upsertIds {
			filters = append(filters, fmt.Sprintf("Id eq %d", id))
		}

		items, err := query.Filter(strings.Join(filters, " or ")).Get()
		if err != nil {
			return o.State.ChangeToken, changesCnt, err
		}

		if err := o.Upsert(ctx, itemsToUpsert(items)); err != nil {
			return o.State.ChangeToken, changesCnt, err
		}
	}

	// Deleted
	var deleteIds []int
	for _, ch := range changes.Data() {
		if ch.ChangeType == deleteChangeType {
			deleteIds = append(deleteIds, ch.ItemID)
		}
	}

	if err := o.Delete(ctx, deleteIds); err != nil {
		return o.State.ChangeToken, changesCnt, err
	}

	return changes.Data()[changesCnt-1].ChangeToken.StringValue, changesCnt, nil
}
