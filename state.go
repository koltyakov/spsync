package spsync

import "time"

// Mode synchronization modes enum
type Mode string

const (
	// Full sync mode
	Full Mode = "Full"
	// Incremental sync mode
	Incr Mode = "Incr"
)

// State structure for defining fault tolerant syncronization state
type State struct {
	// SharePoint List URI, e.g. "Lists/ListA"
	EntID string `json:"entId"`
	// Sync mode: Full or Incr
	SyncMode Mode `json:"syncMode"`
	// Last successful sync timestamp
	SyncDate time.Time `json:"syncDate"`
	// Multistaged sync step
	SyncStage string `json:"syncStage"`
	// List change API token
	ChangeToken string `json:"changeToken"`
	// List pagination skip token
	PageToken string `json:"pageToken"`
	// Number of failed in a row sync sessions
	Fails int `json:"fails"`
}
