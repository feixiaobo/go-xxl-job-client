package utils

import (
	"log"
	"reflect"
)

func ReflectStructToMap(ob interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	object := reflect.ValueOf(ob)
	ref := object.Elem()
	typeOfType := ref.Type()
	if typeOfType.Kind() != reflect.Struct {
		log.Println("Check type error not struct")
		return nil
	}
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		res[typeOfType.Field(i).Name] = field.Interface()
	}
	return res
}
