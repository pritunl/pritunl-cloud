package relations

import (
	"fmt"
	"reflect"
	"strings"
)

type Response struct {
	Id        any
	Label     string
	Fields    []Field
	Relations []Related
}

type Related struct {
	Label     string
	Resources []Resource
}

type Resource struct {
	Id        any
	Type      string
	Fields    []Field
	Relations []Related
}

type Field struct {
	Key   string
	Label string
	Value any
}

func (r *Response) Yaml() string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("ID: %v\n", r.Id))
	output.WriteString(fmt.Sprintf("Label: %s\n", r.Label))

	if len(r.Fields) > 0 {
		output.WriteString("Fields:\n")
		for _, field := range r.Fields {
			output.WriteString(fmt.Sprintf(
				"  %s: %s\n",
				field.Label,
				field.yaml(),
			))
		}
	}

	if len(r.Relations) > 0 {
		output.WriteString("Relations:\n")
		for _, rel := range r.Relations {
			for _, resource := range rel.Resources {
				output.WriteString(resource.yaml(0))
			}
		}
	}

	return output.String()
}

func (r Resource) yaml(indent int) string {
	var output strings.Builder
	indentStr := strings.Repeat(" ", indent)

	output.WriteString(fmt.Sprintf("%s- ID: %v\n", indentStr, r.Id))
	output.WriteString(fmt.Sprintf("%s  Type: %s\n", indentStr, r.Type))

	if len(r.Fields) > 0 {
		output.WriteString(fmt.Sprintf("%s  Fields:\n", indentStr))
		for _, field := range r.Fields {
			output.WriteString(fmt.Sprintf(
				"%s    %s: %s\n",
				indentStr,
				field.Label,
				field.yaml(),
			))
		}
	}

	if len(r.Relations) > 0 {
		output.WriteString(fmt.Sprintf("%s  Relations:\n", indentStr))
		for _, rel := range r.Relations {
			for _, resource := range rel.Resources {
				output.WriteString(resource.yaml(indent + 2))
			}
		}
	}

	return output.String()
}

func (f Field) yaml() string {
	if f.Value == nil {
		return "null"
	}

	v := reflect.ValueOf(f.Value)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		var items []string
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			items = append(items, fmt.Sprintf("%v", item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	}

	if v.Kind() == reflect.String {
		s := f.Value.(string)
		if strings.ContainsAny(s, ":#{}[]&*!|>'\"\n") {
			return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
		}
		return s
	}

	return fmt.Sprintf("%v", f.Value)
}
