package treeprint

import (
	"reflect"
	"strings"
)

func tagSpec(tag string) (name string, omit bool) {
	parts := strings.Split(tag, ",")
	if len(parts) < 2 {
		return tag, false
	}
	if parts[1] == "omitempty" {
		return parts[0], true
	}
	return parts[0], false
}

func filterTags(tag reflect.StructTag) string {
	tags := strings.Split(string(tag), " ")
	filtered := make([]string, 0, len(tags))
	for i := range tags {
		if strings.HasPrefix(tags[i], "tree:") {
			continue
		}
		filtered = append(filtered, tags[i])
	}
	return strings.Join(filtered, " ")
}
