package spsync

// Events hook events during sync session
// usefull for logging in unopinionated way
type Events struct {
	FullSyncStarted   func(entity string, isBlank bool)
	FullSyncCompleted func(entity string, isBlank bool)
	FullSyncRequest   func(entity string, query string)
	IncrSyncStarted   func(entity string)
	IncrSyncCompleted func(entity string)
	IncrSyncRequest   func(entity string, startToken string, endToken string)
}

// Extends events handlers with blank stubs
func ensureEvents(e *Events) *Events {
	if e == nil {
		e = &Events{}
	}

	if e.FullSyncStarted == nil {
		e.FullSyncStarted = func(entity string, isBlank bool) {}
	}

	if e.FullSyncCompleted == nil {
		e.FullSyncCompleted = func(entity string, isBlank bool) {}
	}

	if e.FullSyncRequest == nil {
		e.FullSyncRequest = func(entity string, query string) {}
	}

	if e.IncrSyncStarted == nil {
		e.IncrSyncStarted = func(entity string) {}
	}

	if e.IncrSyncCompleted == nil {
		e.IncrSyncCompleted = func(entity string) {}
	}

	if e.IncrSyncRequest == nil {
		e.IncrSyncRequest = func(entity string, startToken string, endToken string) {}
	}

	return e
}
