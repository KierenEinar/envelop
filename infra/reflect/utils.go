package reflect

import "reflect"

func Struct2Map (obj interface{}) map[string]interface{} {
	t:=reflect.TypeOf(obj)
	v:=reflect.ValueOf(obj)
	m:=make (map[string]interface{})
	for i:=0; i<t.NumField(); i++ {
		m[t.Field(i).Name] = v.Field(i).Interface()
	}
	return m
}

