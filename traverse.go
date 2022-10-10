package snapsdb

import (
	"fmt"
	"reflect"
	"time"

	"github.com/vblegend/snapsdb/util"
)

func Traverse(object any) {
	rv := util.Indirect(reflect.ValueOf(object))
	val := rv.Interface()

	switch oType := val.(type) {
	case string:

	case bool:

	case int:

	case uint:

	case time.Time:

	case int16:
	case uint16:

	case int32:
	case uint32:

	case int64:
	case uint64:

	case float32:
	case float64:

	case []byte:
		TraverseBinary(oType)
	case ValuePair:
		TraverseMap(oType)
	case map[string]interface{}:
		TraverseMap(oType)
	default:

		fmt.Println(rv.Kind())

		switch rv.Kind() {
		case reflect.Struct:
			TraverseStruct(rv)
		case reflect.Slice:
			f := reflect.TypeOf(object)
			fmt.Println(f.Elem().Kind(), "切片数组")
		default:

		}
	}
}

func TraverseBinary(value []byte) {
	fmt.Println("Binary切片数组")
}

func TraverseMap(mv map[string]interface{}) {
	for name, v := range mv {
		fmt.Print("<", name, ">")
		fmt.Println(v)

		Traverse(v)
		// if err := encodeVal(buf, catpath(path, name), name, v); err != nil {
		// 	return nil, err
		// }
	}

}

func TraverseStruct(v reflect.Value) {
	fmt.Println(v.Type().Name(), v.Kind())
	structType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sv := structType.Field(i)
		if sv.PkgPath != "" {
			// Unexported field.
			continue
		}
		name := sv.Name
		// fv := rv.Field(i)
		fv := util.Indirect(v.Field(i))
		fmt.Print("<", name, ">")
		fmt.Printf("[%v]%v\n", fv.Kind(), fv)
		// fmt.Println(name, fv)
		Traverse(fv.Interface())
	}
}
