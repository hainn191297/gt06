package common

import (
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

func ParseBsonMReflect(data bson.M, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("output parameter must be a non-nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Map {
		return errors.New("output parameter must be a map")
	}

	if elem.IsNil() {
		elem.Set(reflect.MakeMap(elem.Type()))
	}

	for k, v := range data {
		key := reflect.ValueOf(k)
		value := reflect.ValueOf(v)

		// Handle nested bson.M or slices recursively
		if nested, ok := v.(bson.M); ok {
			nestedMap := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(any(nil))))
			if err := ParseBsonMReflect(nested, nestedMap.Addr().Interface()); err != nil {
				return err
			}
			value = nestedMap
		} else if slice, ok := v.([]any); ok {
			value = reflect.ValueOf(parseSliceReflect(slice))
		}

		// Set the value in the map
		elem.SetMapIndex(key, value)
	}

	return nil
}

// Helper to process slices recursively
func parseSliceReflect(slice []any) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		if nested, ok := v.(bson.M); ok {
			nestedMap := make(map[string]any)
			_ = ParseBsonMReflect(nested, &nestedMap)
			result[i] = nestedMap
		} else {
			result[i] = v
		}
	}
	return result
}
