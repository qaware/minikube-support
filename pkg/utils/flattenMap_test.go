// This is an adopted version of https://github.com/hashicorp/terraform/blob/master/flatmap/flatten.go
// As the original version it is licensed under the Mozilla Public License 2.0.

package utils

import (
	"reflect"
	"testing"
)

func Test1Flatten(t *testing.T) {
	cases := []struct {
		Input   map[string]interface{}
		Output  map[string]string
		wantErr bool
	}{
		{
			Input: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Output: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
			wantErr: false,
		},

		{
			Input: map[string]interface{}{
				"foo": []string{
					"one",
					"two",
				},
			},
			Output: map[string]string{
				"foo[0]": "one",
				"foo[1]": "two",
			},
			wantErr: false,
		},

		{
			Input: map[string]interface{}{
				"foo": []map[interface{}]interface{}{
					{
						"name":    "bar",
						"port":    3000,
						"enabled": true,
					},
				},
			},
			Output: map[string]string{
				"foo[0].name":    "bar",
				"foo[0].port":    "3000",
				"foo[0].enabled": "true",
			},
			wantErr: false,
		},

		{
			Input: map[string]interface{}{
				"foo": []map[interface{}]interface{}{
					{
						"name": "bar",
						"ports": []string{
							"1",
							"2",
						},
					},
				},
			},
			Output: map[string]string{
				"foo[0].name":     "bar",
				"foo[0].ports[0]": "1",
				"foo[0].ports[1]": "2",
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		actual, err := Flatten(tt.Input)
		if (err != nil) != tt.wantErr {
			t.Errorf("Flatten() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		if !reflect.DeepEqual(actual, tt.Output) {
			t.Fatalf(
				"Input:\n\n%#v\n\nOutput:\n\n%#v\n\nExpected:\n\n%#v\n",
				tt.Input,
				actual,
				tt.Output)
		}
	}
}
