package shactions

import (
	"reflect"
)

// intersectMap returns the intersection of both maps: where the values strictly equal.
func intersectMap(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	out := make(map[string]interface{})
	for k, va := range a {
		if vb, ok := b[k]; ok && reflect.DeepEqual(va, vb) {
			out[k] = va
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
