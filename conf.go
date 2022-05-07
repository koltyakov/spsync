package spsync

import "time"

// Default page size of lists queries during full sync
const deafultPageSize = 1000

// Required fields which are always added for EntConf.Select
var requiredFields = []string{"Id", "AuthorId", "EditorId", "Created", "Modified"}

// EntConf List entity configuration
type EntConf struct {
	// OData fields to select from a list
	Select []string
	// OData props to expand in a list
	Expand []string
	// Custom .Top() value, default is 1000
	Top int
}

// ListItem abstract item structure
type ListItem struct {
	// SharePoint List item ID
	ID int
	// Author user ID
	AuthorID int
	// Created at timestamp
	Created time.Time
	// Editor user ID
	EditorID int
	// Modified at timestamp
	Modified time.Time
	// Item version
	Version int
	// SharePoint List item metadata
	Data map[string]interface{}
}
