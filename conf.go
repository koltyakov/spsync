package spsync

// Default page size of lists queries during full sync
const pageSize = 1000

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

// AbsItem abstract item structure
type AbstItem struct {
	// SharePoint List item ID
	ID int
	// SharePoint List item metadata
	Data map[string]interface{}
}
