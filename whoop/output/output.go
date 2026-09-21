package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Mode uint8

const (
	Human Mode = iota
	JSON
	Plain
)

func JSONValue(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}
func Error(w io.Writer, err error) { _, _ = fmt.Fprintln(w, "whoop: "+err.Error()) }

// PlainValue emits deterministic tab-separated scalar fields for scripts that do not want JSON.
func PlainValue(w io.Writer, value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			_, _ = fmt.Fprintln(w, "")
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := range v.Len() {
			if err := plainStruct(w, v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}
	return plainStruct(w, v)
}
func plainStruct(w io.Writer, v reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		_, err := fmt.Fprintln(w, scalar(v))
		return err
	}
	type pair struct{ name, value string }
	var fields []pair
	t := v.Type()
	for i := range v.NumField() {
		f := v.Field(i)
		if !f.CanInterface() {
			continue
		}
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			name = t.Field(i).Name
		}
		if f.Kind() == reflect.Pointer && f.IsNil() {
			continue
		}
		fields = append(fields, pair{name, scalar(f)})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
	values := make([]string, len(fields))
	for i := range fields {
		values[i] = fields[i].name + "=" + fields[i].value
	}
	_, err := fmt.Fprintln(w, strings.Join(values, "\t"))
	return err
}
func scalar(v reflect.Value) string {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		return scalar(v.Elem())
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		b, _ := json.Marshal(v.Interface())
		return string(b)
	}
}
