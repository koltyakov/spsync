package spsync

import (
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

func itemsToUpsert(items api.ItemsResp) []AbstItem {
	var toUpsert []AbstItem
	for _, item := range items.Data() {
		toUpsert = append(toUpsert, AbstItem{
			ID:   item.Data().ID,
			Data: cleanMap(item.ToMap(), []string{"ID", "odata.id", "odata.editLink", "odata.type"}),
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
