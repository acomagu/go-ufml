package ufml

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type Encoder struct {
	w          io.Writer
	hasPadding bool
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:          w,
		hasPadding: false,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	rows, err := toRows(reflect.ValueOf(v))
	if err != nil {
		return err
	}

	sw := toStrWithWidthNotTable(rows)

	for _, row := range sw {
		for ic, col := range row {
			s := col.s
			if e.hasPadding {
				s = putMargin(s, col.w)
			}

			if ic > 0 {
				fmt.Fprint(e.w, " ") // delimiter
			}
			fmt.Fprint(e.w, s)
		}
		fmt.Fprintln(e.w)
	}

	return nil
}

func (e *Encoder) SetJustifyColumn(on bool) {
	e.hasPadding = on
}

type strWithWidth struct {
	s string
	w int
}

func toStrWithWidthNotTable(rows [][]string) [][]strWithWidth {
	var ret [][]strWithWidth

	var width int
	for _, row := range rows {
		if len(row) == 0 {
			for range rows {
				ret = append(ret, []strWithWidth{})
			}
			return ret
		}
		width = maxInt(width, len(row[0]))
	}

	for i := 0; i < len(rows); {
		colv := rows[i][0]
		var subRows [][]string
		for ; i < len(rows); i++ {
			if rows[i][0] != colv {
				break
			}

			subRows = append(subRows, rows[i][1:])
		}

		subrets := toStrWithWidthNotTable(subRows)
		for _, subret := range subrets {
			ret = append(ret, append([]strWithWidth{{s: colv, w: width}}, subret...))
		}
	}
	return ret
}

func toRows(rv reflect.Value) ([][]string, error) {
	var ret [][]string

	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		var err error
		ret, err = toRows(rv.Elem())
		if err != nil {
			return nil, err
		}

	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i)
			prows, err := toRows(item)
			if err != nil {
				return nil, err
			}

			pret := prefixSl(fmt.Sprintf("#%d", i), prows)
			ret = append(ret, pret...)
		}

	case reflect.Map:
		for _, key := range rv.MapKeys() {
			item := rv.MapIndex(key)
			skey, ok := notatePrim(key)
			if !ok {
				return nil, errors.Errorf("unsupported type %t as key of map", key.Type())
			}

			prows, err := toRows(item)
			if err != nil {
				return nil, err
			}

			pret := prefixSl(escape(skey), prows)
			ret = append(ret, pret...)
		}

	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			key := rv.Type().Field(i).Name
			item := rv.Field(i)

			prows, err := toRows(item)
			if err != nil {
				return nil, err
			}

			pret := prefixSl(escape(key), prows)
			ret = append(ret, pret...)
		}

	default:
		if nt, ok := notatePrim(rv); ok {
			ret = [][]string{{nt}}
		} else {
			return nil, errors.Errorf("unsupported type %#+v: %v\n", rv, rv.Kind())
		}
	}

	return ret, nil
}

func notatePrim(rv reflect.Value) (string, bool) {
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return "<null>", true
		}

		return notatePrim(rv.Elem())
	case reflect.String:
		return escape(rv.String()), true
	case reflect.Bool:
		return fmt.Sprintf("<%t>", rv.Bool()), true
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("<%v>", rv.Float()), true
	default:
		return "", false
	}
}

func prefixSl(p string, ss [][]string) [][]string {
	var ret [][]string
	for _, s := range ss {
		var cols []string
		cols = append(cols, p)
		cols = append(cols, s...)
		ret = append(ret, cols)
	}

	return ret
}

func escape(s string) string {
	if len(s) == 0 {
		return "<>"
	}

	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, " ", "\\ ", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	if s[0] == '<' || s[0] == '#' {
		s = "\\" + s
	}

	return s
}

func putMargin(orig string, width int) string {
	return orig + createMargin(width-len(orig))
}

func createMargin(n int) string {
	c := ' '

	var rs []rune
	for i := 0; i < n; i++ {
		rs = append(rs, c)
	}

	return string(rs)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
