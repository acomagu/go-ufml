package ufml

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"text/scanner"
)

type Token interface{}

type Decoder struct {
	s *bufio.Scanner
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		s: bufio.NewScanner(r),
	}
}

func (d *Decoder) Token() (Token, error) {
	json.Decoder
}

func (d *Decoder) decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("invalid unmarshal error: %s", reflect.TypeOf(v))
	}

	d.decodeLine(rv)
}

func (d *Decoder) decodeLine(v *reflect.Value) (interface{}, error) {
	var err error
	switch d.s.Peek() {
	case '<':
		v, err = d.scanBracket(v)
	case '#':
		v, err = d.scanArrIndex()
	default:
		v, err = d.scanStr()
	}
	if err != nil {
		return nil, err
	}

	d.skipBrank()
	if d.Peek() == '\n' {
		d.Next()
		return v, nil
	}

	deeper, err := d.decodeLine()
	if err != nil {
		return nil, err
	}

	return &kv{
		key:   v,
		value: deeper,
	}, nil
}

func (d *Decoder) isBrank(r rune) bool {
	return r == ' '
}

func (d *Decoder) skipBrank() {
	for d.isBrank(d.Peek()) {
		d.Next()
	}
}

func (d *Decoder) isStrChar(r rune) bool {
	return !d.isBrank(r) && r != '\n' && r != scanner.EOF
}

func (d *Decoder) scanStr() (string, error) {
	var s string
	for ; d.isStrChar(d.Peek()); d.Next() {
		if d.Peek() == '\\' {
			d.Next()
			if d.Peek() == 'n' {
				s += "\n"
				continue
			}
		}

		s += string(d.Peek())
	}

	return s, nil
}

func (d *Decoder) scanBracket(v *reflect.Value) error {
	d.Next() // skip `<`

	var err error
	switch {
	case d.isNumChar(d.Peek()):
		err = d.scanNum()
	case d.isKwdChar(d.Peek()):
		err = d.scanKwd()
	case d.Peek() == '>':
		switch v.Kind {
		case reflect.String:
			v.SetString("")
		default:
			return nil, "could not set string"
		}
	default:
		return nil, errors.New("invalid bracket value")
	}
	if err != nil {
		return nil, err
	}

	if d.Peek() != '>' {
		return nil, errors.Errorf("unexpected '%c'", d.Peek())
	}

	d.Next()

	return v, nil
}
