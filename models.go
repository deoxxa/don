package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type UIStatus struct {
	ID          string    `json:"id"`
	AuthorAcct  string    `json:"authorAcct"`
	AuthorName  string    `json:"authorName"`
	Time        time.Time `json:"time"`
	ContentText string    `json:"contentText"`
	ContentHTML string    `json:"contentHTML"`
}

func toPrimitive(v interface{}) interface{} {
	switch e := v.(type) {
	case string, int, int8, int16, int32, uint8, uint16, uint32:
		return e
	}

	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}

		return toPrimitive(val.Elem().Interface())
	case reflect.Map:
		r := make(map[string]interface{})
		for _, k := range val.MapKeys() {
			r[k.String()] = toPrimitive(val.MapIndex(k).Interface())
		}
		return r
	case reflect.Slice:
		r := make([]interface{}, val.Len())
		for i, j := 0, val.Len(); i < j; i++ {
			r[i] = toPrimitive(val.Index(i).Interface())
		}
		return r
	case reflect.Struct:
		var hasValid, hasValue bool

		if val.NumField() == 2 {
			_, hasValid = val.Type().FieldByName("Valid")
			_, hasValue = val.Type().FieldByName("Value")

			if hasValid && hasValue {
				if val.FieldByName("Valid").Bool() == false {
					return nil
				}

				return toPrimitive(val.FieldByName("Value").Interface())
			}
		}

		r := make(map[string]interface{})
		for i, j := 0, val.NumField(); i < j; i++ {
			if val.Field(i).CanInterface() {
				f := val.Type().Field(i)

				k := f.Name
				omitEmpty := false

				if t := f.Tag.Get("json"); t != "" {
					b := strings.Split(t, ",")
					if len(b[0]) != 0 {
						k = b[0]
					}

					for _, s := range b[1:] {
						if s == "omitempty" {
							omitEmpty = true
						}
					}
				}

				if reflect.DeepEqual(reflect.Zero(f.Type).Interface(), val.Field(i).Interface()) && omitEmpty {
					continue
				}

				r[k] = toPrimitive(val.Field(i).Interface())
			}
		}

		if len(r) == 0 {
			return nil
		}

		return r
	}

	if e, ok := v.(fmt.Stringer); ok {
		return e.String()
	}

	return v
}
