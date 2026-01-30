package botzilla

import "reflect"

func registerType[T any]() {
	var zero T
	t := reflect.TypeOf(zero)
	protoFile := ""

}
