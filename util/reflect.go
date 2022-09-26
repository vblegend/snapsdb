package util

import (
	"errors"
	"reflect"
)

// parse map interface typed
// returm [map_pointer,map_type,map_keytype,slice_type,element_type,error]
func ParseMapPointer(key_map interface{}) (*reflect.Value, *reflect.Type, *reflect.Kind, *reflect.Type, *reflect.Type, error) {
	// 获取slice的类型
	// read metainfo
	map_pointer := reflect.ValueOf(key_map)
	if map_pointer.Kind() != reflect.Ptr {
		return nil, nil, nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	// get list typed
	type_interface := reflect.TypeOf(key_map)
	// get list typed pointer typed
	type_map := type_interface.Elem()
	// get element typed
	type_slice := type_map.Elem()
	// get element typed
	type_keys := type_map.Key()
	if type_keys.Kind() != reflect.String && type_keys.Kind() != reflect.Int64 {
		return nil, nil, nil, nil, nil, errors.New("map key must be of type string or int64'")
	}
	type_element := type_slice.Elem()
	type_key := type_keys.Kind()
	return &map_pointer, &type_map, &type_key, &type_slice, &type_element, nil
}

// parse slice interface typed
// returm [slice_pointer,origin_slice,element_type,error]
func ParseSlicePointer(list interface{}, clearList bool) (*reflect.Value, *reflect.Value, *reflect.Type, error) {
	// read metainfo
	slice_pointer := reflect.ValueOf(list)
	if slice_pointer.Kind() != reflect.Ptr {
		return nil, nil, nil, errors.New("invalid argument 'list interface{}'")
	}
	origin_slice := slice_pointer.Elem()
	// get list typed
	type_interface := reflect.TypeOf(list)
	// get list typed pointer typed
	type_slice := type_interface.Elem()
	// get element typed
	element_type := type_slice.Elem()
	if clearList {
		origin_slice = reflect.Zero(origin_slice.Type())
	}
	return &slice_pointer, &origin_slice, &element_type, nil
}
