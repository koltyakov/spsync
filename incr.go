package spsync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/koltyakov/gosip/api"
)

// Incremental synchronization session flow
func incrSyncSession(ctx context.Context, s *Options) (*State, error) {
	syncDate := time.Now()

	state := s.State

	sp := s.SP
	ent := sp.Web().GetList(s.State.EntID)

	tillToken, err := ent.Changes().GetCurrentToken()
	if err != nil {
		return state, err
	}

	completed := false
	for !completed {
		changeToken, changes, err := incrSyncPaged(ctx, ent, state.ChangeToken, tillToken, s.EntConf, s.Upsert, s.Delete)
		if err != nil {
			return state, err
		}
		state.ChangeToken = changeToken
		if changes == 0 {
			completed = true
		}
	}

	// Success completion state update
	state.PageToken = ""
	state.Fails = 0
	state.SyncDate = syncDate
	state.ChangeToken = tillToken
	state.SyncStage = ""

	return state, nil
}

// Change API paged responce processing
func incrSyncPaged(ctx context.Context, e *api.List, startToken string, endToken string, c *EntConf, up UpsertHandler, del DeleteHandler) (string, int, error) {
	// Default 100 items per page is used for getting changes
	// the page size increase is not recommended as change API returns only IDs
	// and when IDs are used to construct requests to get specific items
	changes, _ := e.Changes().Top(100).GetChanges(&api.ChangeQuery{
		ChangeTokenStart: startToken,
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
		query := e.Items().Conf(api.HeadersPresets.Minimalmetadata).Top(len(upsertIds))
		query = appendOData(query, c)

		var filters []string
		for _, id := range upsertIds {
			filters = append(filters, fmt.Sprintf("Id eq %d", id))
		}

		items, err := query.Filter(strings.Join(filters, " or ")).Get()
		if err != nil {
			return startToken, changesCnt, err
		}

		if err := up(ctx, itemsToUpsert(items)); err != nil {
			return startToken, changesCnt, err
		}
	}

	// Deleted
	var deleteIds []int
	for _, ch := range changes.Data() {
		if ch.ChangeType == deleteChangeType {
			deleteIds = append(deleteIds, ch.ItemID)
		}
	}

	if err := del(ctx, deleteIds); err != nil {
		return startToken, changesCnt, err
	}

	return changes.Data()[changesCnt-1].ChangeToken.StringValue, changesCnt, nil
}
