package spsync

import (
	"strconv"
	"strings"

	"github.com/koltyakov/gosip/api"
)

func appendOData(query *api.Items, conf *EntConf) *api.Items {
	if conf.Select != nil && len(conf.Select) > 0 {
		for _, field := range requiredFields {
			if !contains(conf.Select, field) {
				conf.Select = append(conf.Select, field)
			}
		}
		query = query.Select(strings.Join(conf.Select, ","))
	}
	if conf.Expand != nil && len(conf.Expand) > 0 {
		query = query.Expand(strings.Join(conf.Expand, ","))
	}
	return query
}

func itemsToUpsert(items api.ItemsResp) []ListItem {
	var toUpsert []ListItem
	for _, item := range items.Data() {
		d := item.Data()
		m := item.ToMap()
		version, _ := strconv.Atoi(m["odata.etag"].(string))
		toUpsert = append(toUpsert, ListItem{
			ID:       d.ID,
			AuthorID: d.AuthorID,
			EditorID: d.EditorID,
			Created:  d.Created,
			Modified: d.Modified,
			Version:  version,
			Data:     cleanMap(item.ToMap(), []string{"ID", "odata.id", "odata.editLink", "odata.type", "odata.etag"}),
		})
	}
	return toUpsert
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func cleanMap(m map[string]interface{}, clean []string) map[string]interface{} {
	for _, field := range clean {
		delete(m, field)
	}
	return m
}
