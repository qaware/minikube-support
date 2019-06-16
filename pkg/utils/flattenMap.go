// This is an adopted version of https://github.com/hashicorp/terraform/blob/master/flatmap/flatten.go
// As the original version it is licensed under the Mozilla Public License 2.0.

package utils

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"reflect"
)

// Flatten takes a structure and turns into a flat map[string]string.
//
// Within the "thing" parameter, only primitive values are allowed. Structs are
// not supported. Therefore, it can only be slices, maps, primitives, and
// any combination of those together.
//
// See the tests for examples of what inputs are turned into.
func Flatten(thing map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string)
	var errors *multierror.Error
	for k, raw := range thing {
		errors = multierror.Append(errors, flatten(result, k, reflect.ValueOf(raw)))
	}

	return result, errors.ErrorOrNil()
}

func flatten(result map[string]string, prefix string, v reflect.Value) error {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			result[prefix] = "true"
		} else {
			result[prefix] = "false"
		}
	case reflect.Int:
		result[prefix] = fmt.Sprintf("%d", v.Int())
	case reflect.Map:
		return flattenMap(result, prefix, v)
	case reflect.Slice:
		return flattenSlice(result, prefix, v)
	case reflect.String:
		result[prefix] = v.String()
	default:
		return fmt.Errorf("unknown value: %s", v)
	}
	return nil
}

func flattenMap(result map[string]string, prefix string, v reflect.Value) error {
	var errors *multierror.Error
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}

		if k.Kind() != reflect.String {
			panic(fmt.Sprintf("%s: map key is not string: %s", prefix, k))
		}

		errors = multierror.Append(errors, flatten(result, fmt.Sprintf("%s.%s", prefix, k.String()), v.MapIndex(k)))
	}
	return errors.ErrorOrNil()
}

func flattenSlice(result map[string]string, prefix string, v reflect.Value) error {
	var errors *multierror.Error
	//prefix = prefix + "."
	for i := 0; i < v.Len(); i++ {
		errors = multierror.Append(errors, flatten(result, fmt.Sprintf("%s[%d]", prefix, i), v.Index(i)))
	}
	return errors.ErrorOrNil()
}
