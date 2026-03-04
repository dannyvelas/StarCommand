package helpers

import "maps"

func MergeMaps[K comparable, V any](m1, m2 map[K]V) map[K]V {
	newMap := make(map[K]V)
	maps.Copy(newMap, m1)
	maps.Copy(newMap, m2)
	return newMap
}
