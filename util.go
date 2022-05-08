package spsync

import (
	"strconv"
	"strings"

	"github.com/koltyakov/gosip/api"
)

// Applies entity OData configuration
func appendOData(items *api.Items, ent *Ent) *api.Items {
	if ent.Select != nil && len(ent.Select) > 0 {
		for _, field := range requiredFields {
			if !contains(ent.Select, field) {
				ent.Select = append(ent.Select, field)
			}
		}
		items = items.Select(strings.Join(ent.Select, ","))
	}
	if ent.Expand != nil && len(ent.Expand) > 0 {
		items = items.Expand(strings.Join(ent.Expand, ","))
	}
	return items
}

// Converts API items response to abstract sync items struct
func itemsToUpsert(items api.ItemsResp) []Item {
	var toUpsert []Item
	for _, item := range items.Data() {
		d := item.Data()
		m := item.ToMap()
		version, _ := strconv.Atoi(m["odata.etag"].(string))
		toUpsert = append(toUpsert, Item{
			ID:       d.ID,
			Created:  d.Created,
			Modified: d.Modified,
			Version:  version,
			Data:     cleanMap(item.ToMap(), []string{"ID", "odata.id", "odata.editLink", "odata.type", "odata.etag"}),
		})
	}
	return toUpsert
}

// Checks if a slice contains a string
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Removes provided props from a map
func cleanMap(m map[string]interface{}, clean []string) map[string]interface{} {
	for _, field := range clean {
		delete(m, field)
	}
	return m
}
