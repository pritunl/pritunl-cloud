package resource

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func selector(v interface{}, key string) (val string, err error) {
	valRef := reflect.ValueOf(v)

	if valRef.Kind() != reflect.Ptr || valRef.IsNil() {
		err = &errortypes.ParseError{
			errors.New("Selector input invalid"),
		}
		return
	}

	elm := valRef.Elem()
	if elm.Kind() != reflect.Struct {
		err = &errortypes.ParseError{
			errors.New("Selector kind invalid"),
		}
		return
	}

	typ := elm.Type()

	for i := 0; i < elm.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag == key {
			fieldVal := elm.Field(i)

			if fieldVal.Kind() == reflect.Slice {
				var elements []string

				for j := 0; j < fieldVal.Len(); j++ {
					elements = append(elements,
						selectString(fieldVal.Index(j).Interface()))
				}

				val = strings.Join(elements, ",")
			} else {
				val = selectString(elm.Field(i).Interface())
			}

			return
		}
	}

	return
}

func selectString(obj interface{}) string {
	if oid, ok := obj.(primitive.ObjectID); ok {
		return oid.Hex()
	}

	val := reflect.ValueOf(obj)
	val = reflect.Indirect(val)

	method := val.MethodByName("String")
	if method.IsValid() &&
		method.Type().NumIn() == 0 &&
		method.Type().NumOut() == 1 &&
		method.Type().Out(0).Kind() == reflect.String {

		result := method.Call(nil)
		return result[0].String()
	}

	return fmt.Sprintf("%v", obj)
}
