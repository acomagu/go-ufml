package ufml

import (
	"bytes"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {
	tests := map[string]struct {
		v interface{}

		res     string
		wantErr bool
	}{
		"map": {
			v: map[string]string{
				"a": "b",
				"c": "d",
			},
			res:     "a b\nc d\n",
			wantErr: false,
		},
		"nil map": {
			v:       map[string]string(nil),
			res:     "null",
			wantErr: false,
		},
		"nil pointer": {
			v:       nil,
			res:     "null",
			wantErr: false,
		},
	}

	for name, _test := range tests {
		test := _test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := bytes.NewBuffer(nil)
			enc := NewEncoder(res)

			if err := enc.Encode(test.v); (err != nil) != test.wantErr {
				t.Errorf("Encoder.Encode() error = %v, wantErr %v", err, test.wantErr)
			}
			if res.String() != test.res {
				t.Errorf("Encoder.Encode() wrotes %s but wanted %s", res.String(), test.res)
			}
		})
	}
}
