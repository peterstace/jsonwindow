package jsonwindow

import (
	"fmt"
	"io"
)

func WalkObject(raw []byte, fn func(key, val []byte) error) error {
	i := 0
	i += countWS(raw)
	if i >= len(raw) {
		return io.ErrUnexpectedEOF
	}
	if raw[i] != '{' {
		return fmt.Errorf("JSON object must start with '{'")
	}
	i++
	_, err := continueObject(raw[i:], fn)
	return err
}

func countWS(raw []byte) int {
	i := 0
	for {
		if i >= len(raw) {
			return i
		}
		switch raw[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
}

// assume: raw is non-empty and WS stripped
func parseValue(raw []byte) (int, error) {
	switch raw[0] {
	case '{':
		n, err := continueObject(raw[1:], nil)
		return 1 + n, err
	case '[':
		n, err := continueArray(raw[1:])
		return 1 + n, err
	case '"':
		panic("not implemented yet")
	case 't':
		if len(raw) < len("true") {
			return 0, io.ErrUnexpectedEOF
		}
		if string(raw[:len("true")]) != "true" {
			return 0, fmt.Errorf("'t' must be followed by 'rue'")
		}
		return len("true"), nil
	case 'f':
		if len(raw) < len("false") {
			return 0, io.ErrUnexpectedEOF
		}
		if string(raw[:len("false")]) != "false" {
			return 0, fmt.Errorf("'f' must be followed by 'alse'")
		}
		return len("false"), nil
	case 'n':
		if len(raw) < len("null") {
			return 0, io.ErrUnexpectedEOF
		}
		if string(raw[:len("null")]) != "null" {
			return 0, fmt.Errorf("'n' must be followed by 'ull'")
		}
		return len("null"), nil
	default:
		if raw[0] >= '0' && raw[0] <= '9' || raw[0] == '-' {
			panic("not implemented yet")
		}
		return 0, fmt.Errorf("JSON value must start with" +
			" '{', '[', '\"', 't', 'f', 'n', '-', or a digit")
	}
}

// assume: '{' has already been consumed
func continueObject(raw []byte, fn func(key, val []byte) error) (int, error) {
	// Check for empty object (closing curly):
	i := 0
	i += countWS(raw[i:])
	if i >= len(raw) {
		return 0, io.ErrUnexpectedEOF
	}
	if raw[i] == '}' {
		i++
		return i, nil
	}

	for {
		// Consume key. It must be a string.
		i += countWS(raw[i:])
		key, err := parseString(raw[i:])
		if err != nil {
			return 0, err
		}
		i += len(key)

		// Consume the colon separating the key from the value.
		i += countWS(raw[i:])
		if i >= len(raw) {
			return 0, io.ErrUnexpectedEOF
		}
		if raw[i] != ':' {
			return 0, fmt.Errorf("':' must come after key in JSON object")
		}
		i++

		// Consume the value. It doesn't matter what type of value it is.
		i += countWS(raw[i:])
		valLen, err := parseValue(raw[i:])
		if err != nil {
			return 0, err
		}

		// Use callback if provided.
		if fn != nil {
			if err := fn(key, raw[i:i+valLen]); err != nil {
				return 0, err
			}
		}
		i += valLen

		// Check to see if we're at the end of the object, of if there are
		// more key/value pairs.
		i += countWS(raw[:i])
		if i >= len(raw) {
			return 0, io.ErrUnexpectedEOF
		}
		if raw[i] == ',' {
			i++
			continue
		}
		if raw[i] == '}' {
			i++
			return i, nil
		}
		return 0, fmt.Errorf("',' or '}' must come after value in JSON object")
	}
}

// assume: '[' has already been consumed
// TODO: give callback
func continueArray(raw []byte) (int, error) {
	// Check for empty array (closing square):
	i := 0
	i += countWS(raw[i:])
	if i >= len(raw) {
		return 0, io.ErrUnexpectedEOF
	}
	if raw[i] == ']' {
		i++
		return i, nil
	}

	for {
		// TODO
	}
}

func parseString(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	if raw[0] != '"' {
		return nil, fmt.Errorf("JSON string must start with '\"'")
	}

	i := 1
	for {
		if i >= len(raw) {
			return nil, io.ErrUnexpectedEOF
		}
		c := raw[i]
		i++
		if c == '"' {
			return raw[:i], nil
		}
		if c == '\\' {
			if i >= len(raw) {
				return nil, io.ErrUnexpectedEOF
			}
			if raw[i] == 'u' {
				i += 4 // Skip the next 4 hex digits.
			} else {
				i++ // Skip the single escaped character.
			}
		}
	}
}