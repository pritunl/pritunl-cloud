package statec

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
)

const canonMaxDepth = 32

var (
	canonTimeType     = reflect.TypeOf(time.Time{})
	canonObjectIdType = reflect.TypeOf(bson.ObjectID{})
)

func Canon(val interface{}) string {
	b := &strings.Builder{}
	canonValue(b, reflect.ValueOf(val), 0)
	return b.String()
}

func canonSet(b *strings.Builder, s set.Set, depth int) {
	items := []string{}
	for item := range s.Iter() {
		sb := &strings.Builder{}
		canonValue(sb, reflect.ValueOf(item), depth+1)
		items = append(items, sb.String())
	}
	sort.Strings(items)

	b.WriteString("set[")
	b.WriteString(strings.Join(items, ","))
	b.WriteString("]")
}

func canonValue(b *strings.Builder, v reflect.Value, depth int) {
	if !v.IsValid() {
		b.WriteString("nil")
		return
	}

	if depth > canonMaxDepth {
		b.WriteString("...")
		return
	}

	if v.CanInterface() {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			if v.IsNil() {
				b.WriteString("nil")
				return
			}
		}

		switch val := v.Interface().(type) {
		case set.Set:
			canonSet(b, val, depth)
			return
		case bson.ObjectID:
			b.WriteString(val.Hex())
			return
		case time.Time:
			b.WriteString(val.UTC().Format(time.RFC3339Nano))
			return
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			b.WriteString("nil")
			return
		}
		canonValue(b, v.Elem(), depth+1)

	case reflect.Struct:
		if v.Type() == canonTimeType || v.Type() == canonObjectIdType {
			b.WriteString(fmt.Sprintf("%v", v))
			return
		}

		b.WriteString("{")
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}

			b.WriteString(field.Name)
			b.WriteString(":")
			canonValue(b, v.Field(i), depth+1)
			b.WriteString(",")
		}
		b.WriteString("}")

	case reflect.Map:
		entries := make([]string, 0, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			kb := &strings.Builder{}
			canonValue(kb, iter.Key(), depth+1)
			kb.WriteString("=")
			canonValue(kb, iter.Value(), depth+1)
			entries = append(entries, kb.String())
		}
		sort.Strings(entries)

		b.WriteString("map[")
		b.WriteString(strings.Join(entries, ","))
		b.WriteString("]")

	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.Kind() == reflect.Slice {
				b.WriteString("0x")
				b.WriteString(hex.EncodeToString(v.Bytes()))
				return
			}

			buf := make([]byte, v.Len())
			for i := 0; i < v.Len(); i++ {
				buf[i] = byte(v.Index(i).Uint())
			}
			b.WriteString("0x")
			b.WriteString(hex.EncodeToString(buf))
			return
		}

		b.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			canonValue(b, v.Index(i), depth+1)
			b.WriteString(",")
		}
		b.WriteString("]")

	case reflect.String:
		b.WriteString(strconv.Quote(v.String()))

	case reflect.Bool:
		b.WriteString(strconv.FormatBool(v.Bool()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:

		b.WriteString(strconv.FormatInt(v.Int(), 10))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:

		b.WriteString(strconv.FormatUint(v.Uint(), 10))

	case reflect.Float32, reflect.Float64:
		b.WriteString(strconv.FormatFloat(v.Float(), 'g', -1, 64))

	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		b.WriteString("<")
		b.WriteString(v.Kind().String())
		b.WriteString(">")

	default:
		b.WriteString(fmt.Sprintf("%v", v))
	}
}

func canonList(val interface{}) (items []string) {
	items = []string{}

	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		items = append(items, Canon(val))
		return
	}

	for i := 0; i < v.Len(); i++ {
		b := &strings.Builder{}
		canonValue(b, v.Index(i), 0)
		items = append(items, b.String())
	}

	sort.Strings(items)

	return
}

func canonKey(v reflect.Value) string {
	if v.CanInterface() {
		switch val := v.Interface().(type) {
		case bson.ObjectID:
			return val.Hex()
		case string:
			return val
		}
	}

	return Canon(v.Interface())
}
