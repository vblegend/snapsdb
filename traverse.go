package snapsdb

import (
	"fmt"
	"reflect"
	"time"

	"github.com/vblegend/snapsdb/util"
)

func Traverse(object any, name string) {
	rv := util.Indirect(reflect.ValueOf(object))
	val := rv.Interface()

	switch oType := val.(type) {
	case string:
		WriteVlaue(name, oType)
	case bool:
		WriteVlaue(name, oType)
	case int:
		WriteVlaue(name, oType)
	case uint:
		WriteVlaue(name, oType)
	case time.Time:
		WriteVlaue(name, oType.Unix())
	case int16:
		WriteVlaue(name, oType)
	case uint16:
		WriteVlaue(name, oType)
	case int32:
		WriteVlaue(name, oType)
	case uint32:
		WriteVlaue(name, oType)
	case int64:
		WriteVlaue(name, oType)
	case uint64:
		WriteVlaue(name, oType)
	case float32:
		WriteVlaue(name, oType)
	case float64:
		WriteVlaue(name, oType)
	case []byte:
		TraverseBinary(oType, name)
	case TagPair:
		TraverseMap(oType, name)
	case ValuePair:
		TraverseMap(oType, name)
	default:
		switch rv.Kind() {
		case reflect.Struct:
			TraverseStruct(rv, name)
		case reflect.Slice:
			f := reflect.TypeOf(object)
			fmt.Errorf("%s%s", f.Elem().Kind(), "切片")
			WriteVlaue(name, object)
		default:

		}
	}
}

func WriteVlaue(name string, value any) {
	if name == "" {
		fmt.Println(value)
	} else {
		fmt.Printf("%s: %v\n", name, value)
	}

}

func TraverseBinary(value []byte, name string) {
	WriteVlaue(name, value)
}

func TraverseMap(mv map[string]interface{}, name string) {
	fmt.Println()
	if name != "" {
		fmt.Printf("%s: {\n", name)
	}
	for _name, v := range mv {
		Traverse(v, _name)
	}
	if name != "" {
		fmt.Printf("}\n")
	}
}

func TraverseStruct(v reflect.Value, name string) {
	// fmt.Println(v.Type().Name(), v.Kind())
	if name != "" {
		fmt.Printf("%s: {\n", name)
	}
	structType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sv := structType.Field(i)
		if sv.PkgPath != "" {
			// Unexported field.
			continue
		}
		_name := sv.Name
		// fv := rv.Field(i)
		fv := util.Indirect(v.Field(i))
		// fmt.Print("[", _name, "]")
		Traverse(fv.Interface(), _name)
	}
	if name != "" {
		fmt.Printf("}\n")
	}
}
