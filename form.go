package reason

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type formField struct {
	name  string
	typ   reflect.Type
	index []int
}

func (s *Server) getSchemaFields(t reflect.Type) ([]formField, error) {
	s.formCacheLock.RLock()
	fields := s.formCache[t]
	s.formCacheLock.RUnlock()
	if fields != nil {
		return fields, nil
	}

	fields = make([]formField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sfield := t.Field(i)
		field := formField{}
		tag := sfield.Tag.Get("json")
		if tag != "" {
			if idx := strings.Index(tag, ","); idx != -1 {
				field.name = tag[0:idx]
			} else {
				field.name = tag
			}
		} else {
			field.name = sfield.Name
		}
		field.typ = sfield.Type
		field.index = sfield.Index
		fields = append(fields, field)
	}

	s.formCacheLock.Lock()
	s.formCache[t] = fields
	s.formCacheLock.Unlock()

	return fields, nil
}

func (s *Server) parseForm(r *http.Request, schema interface{}) (interface{}, error) {
	t := reflect.TypeOf(schema)
	fields, err := s.getSchemaFields(t)
	if err != nil {
		return nil, err
	}

	// Create a new instance to write to
	val := reflect.New(t).Elem()

	for _, field := range fields {
		formval := r.FormValue(field.name)

		// Ignore empty values and let the resource handler validate
		if formval != "" {
			switch field.typ.Kind() {
			case reflect.String:
				val.FieldByIndex(field.index).SetString(formval)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intval, err := strconv.ParseInt(formval, 10, 64)
				if err != nil {
					return nil, err
				}
				val.FieldByIndex(field.index).SetInt(intval)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				uintval, err := strconv.ParseUint(formval, 10, 64)
				if err != nil {
					return nil, err
				}
				val.FieldByIndex(field.index).SetUint(uintval)
			case reflect.Float32, reflect.Float64:
				floatval, err := strconv.ParseFloat(formval, 64)
				if err != nil {
					return nil, err
				}
				val.FieldByIndex(field.index).SetFloat(floatval)
			case reflect.Bool:
				boolval := formval == "true" || formval == "1"
				val.FieldByIndex(field.index).SetBool(boolval)
			}
		}
	}

	return val.Interface(), nil
}
