package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func selector(v interface{}, key string, isJson bool) (val string, err error) {
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

	jsonKey := ""
	if isJson {
		keys := strings.SplitN(key, ".", 2)
		if len(keys) == 2 {
			key = keys[0]
			jsonKey = keys[1]
		} else {
			isJson = false
		}
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

				if isJson && jsonKey != "" {
					var jsonData map[string]any
					err = json.Unmarshal([]byte(val), &jsonData)
					if err != nil {
						val = ""
						return
					}

					jsonValue, exists := jsonData[jsonKey]
					if !exists {
						val = ""
						return
					}

					val = jsonValString(jsonValue)
				}
			}

			return
		}
	}

	return
}

func jsonValString(value any) string {
	switch val := value.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case nil:
		return ""
	default:
		jsonBytes, _ := json.Marshal(val)
		return string(jsonBytes)
	}
}

func selectString(obj interface{}) string {
	if oid, ok := obj.(bson.ObjectID); ok {
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
