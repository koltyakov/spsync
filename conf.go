package spsync

import "time"

// Default page size of lists queries during full sync
const defaultPageSize = 1000

// Required fields which are always added for EntConf.Select
var requiredFields = []string{"Id", "Created", "Modified"}

// Ent List entity configuration
type Ent struct {
	// OData fields to select from a list
	Select []string
	// OData props to expand in a list
	Expand []string
	// Custom .Top() value, default is 1000
	Top int
}

// Item abstract item structure
type Item struct {
	// SharePoint List item ID
	ID int
	// Created at timestamp
	Created time.Time
	// Modified at timestamp
	Modified time.Time
	// Item version
	Version int
	// SharePoint List item metadata
	Data map[string]interface{}
}
