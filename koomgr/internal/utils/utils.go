package utils

import "github.com/golang-collections/collections/set"

func Set2stringSlice(set *set.Set) []string {
	keys := make([]string, 0, set.Len())
	set.Do(func(s interface{}) {
		keys = append(keys, s.(string))
	})
	return keys
}
