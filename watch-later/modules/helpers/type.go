package helpers

import (
	"reflect"
)

func IsKind(what interface{}, kind reflect.Kind) bool {
	return reflect.ValueOf(what).Kind() == kind
}
