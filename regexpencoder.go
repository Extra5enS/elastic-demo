package main

import (
	"io"
)

// code is based on json.Encoder source code
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.5:src/encoding/json/stream.go;l=201

type RegexEncode struct {
	w   io.Writer
	err error
}

func NewRegexEncode(inv io.Writer) *RegexEncode {
	return &RegexEncode{inv, nil}
}

func (enc *RegexEncode) Error() error {
	return enc.err
}

const READ_SIZE = 100

// Originaly `v` was interface{}
func (enc *RegexEncode) Encode(v io.Reader) error {
	b := make([]byte, READ_SIZE)
	var (
		n   int
		err error
	)
	for n, err = v.Read(b); err != io.EOF; n, err = v.Read(b) {
		if err != nil {
			break
		}
		offset := 0
		for i, sym := range b {
			switch {
			case sym == '*':
				if _, err := enc.w.Write(b[offset:i]); err != nil {
					enc.err = err
					return err
				}
				offset = i // b[i] = "*", so next Write will put it in.
				if _, err := enc.w.Write([]byte{'.'}); err != nil {
					enc.err = err
					return err
				}
			case isSpecSym(sym):
				if _, err := enc.w.Write(b[offset:i]); err != nil {
					enc.err = err
					return err
				}
				offset = i
				if _, err := enc.w.Write([]byte{'\\'}); err != nil {
					enc.err = err
					return err
				}
			default:
				// do nothing
			}
		}
		enc.w.Write(b[offset:n])
	}

	if err == io.EOF {
		err = nil
	}
	enc.err = err
	return err
}

func isSpecSym(sym byte) bool {
	for _, sp := range []byte{'.', '?', '+', '\\', '{', '}', '[', ']', '(', ')', '"', '|', '#', '@', '&', '<', '>', '~'} {
		if sym == sp {
			return true
		}
	}
	return false
}

// now we can choose only next regexps *<word>, *<word>*, <word>*
func isAccepterRegexp(regexp string) bool {
	for i, sym := range regexp {
		if sym == '*' && (i != 0 || i != len(regexp)-1) {
			return false
		}
	}
	return true
}
