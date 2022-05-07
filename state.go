package spsync

import "time"

// SyncMode synchronization modes enum
type SyncMode string

const (
	// Full sync mode
	FullSync SyncMode = "Full"
	// Incremental sync mode
	IncrSync SyncMode = "Incr"
)

// SyncState structure for defining fault tolerant syncronization state
type SyncState struct {
	// SharePoint List URI, e.g. "Lists/ListA"
	EntID string `json:"entId"`
	// Sync mode: Full or Incr
	SyncMode SyncMode `json:"syncMode"`
	// Last successful sync timestamp
	SyncDate time.Time `json:"syncDate"`
	// Multistaged sync step
	SyncStage string `json:"syncStage"`
	// List change API token
	ChngToken string `json:"changeToken"`
	// List pagination skip token
	SkipToken string `json:"skipToken"`
	// Number of failed in a row sync sessions
	Fails int `json:"fails"`
}
