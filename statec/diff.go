package statec

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/mongo-go-driver/v2/bson"
)

const diffPathLimit = 8

type pathDiffer struct {
	paths     []string
	truncated bool
}

func (d *pathDiffer) full() bool {
	return len(d.paths) >= diffPathLimit
}

func (d *pathDiffer) add(path, note string) {
	if d.full() {
		d.truncated = true
		return
	}

	if path == "" {
		path = "value"
	}
	if note != "" {
		path += " (" + note + ")"
	}

	d.paths = append(d.paths, path)
}

func joinPath(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func indexPath(path, index string) string {
	return path + "[" + index + "]"
}

func canonString(v reflect.Value) string {
	b := &strings.Builder{}
	canonValue(b, v, 0)
	return b.String()
}

func (d *pathDiffer) walk(a, b reflect.Value, path string, depth int) {
	if d.full() {
		d.truncated = true
		return
	}

	if !a.IsValid() || !b.IsValid() {
		if a.IsValid() != b.IsValid() {
			d.add(path, "missing")
		}
		return
	}

	if a.Type() != b.Type() {
		d.add(path, "type")
		return
	}

	if depth > canonMaxDepth {
		return
	}

	if a.CanInterface() && b.CanInterface() {
		switch a.Interface().(type) {
		case set.Set, bson.ObjectID, time.Time:
			if canonString(a) != canonString(b) {
				d.add(path, "")
			}
			return
		}
	}

	switch a.Kind() {
	case reflect.Ptr, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				d.add(path, "nil")
			}
			return
		}
		d.walk(a.Elem(), b.Elem(), path, depth+1)

	case reflect.Struct:
		typ := a.Type()
		for i := 0; i < a.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}

			d.walk(a.Field(i), b.Field(i),
				joinPath(path, field.Name), depth+1)
			if d.full() {
				return
			}
		}

	case reflect.Map:
		valsA := map[string]reflect.Value{}
		valsB := map[string]reflect.Value{}
		if !a.IsNil() {
			iter := a.MapRange()
			for iter.Next() {
				valsA[canonKey(iter.Key())] = iter.Value()
			}
		}
		if !b.IsNil() {
			iter := b.MapRange()
			for iter.Next() {
				valsB[canonKey(iter.Key())] = iter.Value()
			}
		}

		keys := map[string]struct{}{}
		for key := range valsA {
			keys[key] = struct{}{}
		}
		for key := range valsB {
			keys[key] = struct{}{}
		}

		sortedKeys := make([]string, 0, len(keys))
		for key := range keys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {
			valA, okA := valsA[key]
			valB, okB := valsB[key]
			switch {
			case !okA:
				d.add(indexPath(path, key), "statec only")
			case !okB:
				d.add(indexPath(path, key), "state only")
			default:
				d.walk(valA, valB, indexPath(path, key), depth+1)
			}
			if d.full() {
				return
			}
		}

	case reflect.Slice, reflect.Array:
		if a.Type().Elem().Kind() == reflect.Uint8 {
			if canonString(a) != canonString(b) {
				d.add(path, "")
			}
			return
		}

		lenA := a.Len()
		lenB := b.Len()
		if lenA != lenB {
			d.add(path, fmt.Sprintf("len %d != %d", lenA, lenB))
		}

		n := lenA
		if lenB < n {
			n = lenB
		}
		for i := 0; i < n; i++ {
			d.walk(a.Index(i), b.Index(i),
				indexPath(path, strconv.Itoa(i)), depth+1)
			if d.full() {
				return
			}
		}

	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return

	default:
		if canonString(a) != canonString(b) {
			d.add(path, "")
		}
	}
}

func fieldDiffs(a, b interface{}) (paths []string, truncated bool) {
	d := &pathDiffer{}
	d.walk(reflect.ValueOf(a), reflect.ValueOf(b), "", 0)
	paths = d.paths
	truncated = d.truncated
	return
}

func fieldsDetail(a, b interface{}) string {
	paths, truncated := fieldDiffs(a, b)
	if len(paths) == 0 {
		return ""
	}

	detail := strings.Join(paths, ", ")
	if truncated {
		detail += ", ..."
	}

	return detail
}

func elemId(v reflect.Value) (id string, ok bool) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	field := v.FieldByName("Id")
	if !field.IsValid() || !field.CanInterface() {
		return
	}

	objId, isObjId := field.Interface().(bson.ObjectID)
	if !isObjId {
		return
	}

	id = objId.Hex()
	ok = true
	return
}

func listById(val interface{}) (items map[string]reflect.Value, ok bool) {
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
		return
	}

	items = map[string]reflect.Value{}
	for i := 0; i < v.Len(); i++ {
		id, idOk := elemId(v.Index(i))
		if !idOk {
			return nil, false
		}

		if _, dup := items[id]; dup {
			return nil, false
		}

		items[id] = v.Index(i)
	}

	ok = true
	return
}
