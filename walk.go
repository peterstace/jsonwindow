package jsonwindow

import (
	"fmt"
	"io"
)

func WalkObject(raw []byte, fn func(key, val []byte) error) error {
	i := countWS(raw)
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

// countWS counts the number of chars making the whitespace prefix of raw.
//
// Note on whitespace handling:
//
//  - Whitespace should be consumed after each token/structure rather than
//  before each token/structure.
//
//  - The exception is handling whitespace at the start of any exported
//  function.
//
//  - functions with `parse` or `continue` prefixes don't strip whitespace at
//  the end.
//
// TODO: use "outlining"
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

// assume: raw has WS stripped
func parseValue(raw []byte) (int, error) {
	// TODO: use if statements rather than case to be more explicit about
	// ordering?
	// TODO: could use LUT rather than switch?
	// TODO: re-order cases for best speed
	if len(raw) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	switch raw[0] {
	case '{':
		n, err := continueObject(raw[1:], nil)
		return 1 + n, err
	case '[':
		n, err := continueArray(raw[1:])
		return 1 + n, err
	case '"':
		// TODO: continueString rather than parseString
		str, err := parseString(raw)
		return len(str), err
	case 't':
		err := parseExact("rue", raw[1:])
		return len("true"), err
	case 'f':
		err := parseExact("alse", raw[1:])
		return len("false"), err
	case 'n':
		err := parseExact("ull", raw[1:])
		return len("null"), err
	default:
		// TODO: could put number start chars in their own case?
		if isNumberStartChar(raw[0]) {
			n := continueNumber(raw[1:])
			return n + 1, nil
		}
		return 0, fmt.Errorf("JSON value must start with" +
			" '{', '[', '\"', 't', 'f', 'n', '-', or a digit")
	}
}

// assume: '{' has already been consumed
// assume: whitespace _not_ stripped after the '{'
func continueObject(raw []byte, fn func(key, val []byte) error) (int, error) {
	// Check for empty object (closing curly):
	i := countWS(raw)
	if i >= len(raw) {
		return 0, io.ErrUnexpectedEOF
	}
	if raw[i] == '}' {
		i++
		return i, nil
	}

	for {
		// Consume key. It must be a string.
		key, err := parseString(raw[i:])
		if err != nil {
			return 0, err
		}
		i += len(key)
		i += countWS(raw[i:])

		// Consume the colon separating the key from the value.
		if i >= len(raw) {
			return 0, io.ErrUnexpectedEOF
		}
		if raw[i] != ':' {
			return 0, fmt.Errorf("':' must come after key in JSON object")
		}
		i++
		i += countWS(raw[i:])

		// Consume the value. It doesn't matter what type of value it is.
		val, err := parseValue(raw[i:])
		if err != nil {
			return 0, err
		}
		rawVal := raw[i : i+val]
		i += val
		i += countWS(raw[i:])

		// Use callback if provided.
		if fn != nil {
			if err := fn(key, rawVal); err != nil {
				return 0, err
			}
		}

		// Check to see if we're at the end of the object, of if there are
		// more key/value pairs.
		if i >= len(raw) {
			return 0, io.ErrUnexpectedEOF
		}
		if raw[i] == ',' {
			i++
			i += countWS(raw[i:])
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
func continueArray(raw []byte) (int, error) {
	// Check for empty array (closing square):
	i := countWS(raw)
	if i >= len(raw) {
		return 0, io.ErrUnexpectedEOF
	}
	if raw[i] == ']' {
		i++
		return i, nil
	}

	for {
		// Consume the value. It doesn't matter what type of value it is.
		val, err := parseValue(raw[i:])
		if err != nil {
			return 0, err
		}
		i += val
		i += countWS(raw[i:])

		// TODO: give callback

		// Check to see if we're at the end of the array, of if there are more
		// values.
		if i >= len(raw) {
			return 0, io.ErrUnexpectedEOF
		}
		if raw[i] == ',' {
			i++
			i += countWS(raw[i:])
			continue
		}
		if raw[i] == ']' {
			i++
			return i, nil
		}
		return 0, fmt.Errorf("',' or ']' must come after value in JSON array")
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

// assume: first char of number is already consumed
func continueNumber(raw []byte) int {
	i := 0
	for i < len(raw) && isNumberContinueChar(raw[i]) {
		i++
	}
	return i
}

func isNumberStartChar(c byte) bool {
	return false ||
		(c >= '0' && c <= '9') ||
		c == '-'
}

func isNumberContinueChar(c byte) bool {
	return false ||
		(c >= '0' && c <= '9') ||
		c == '.' ||
		c == 'e' ||
		c == 'E' ||
		c == '-' ||
		c == '+'
}

// parseExact checks if the raw has the exact given prefix.
func parseExact(exact string, raw []byte) error {
	if len(raw) < len(exact) {
		return io.ErrUnexpectedEOF
	}
	if string(raw[:len(exact)]) != exact {
		return fmt.Errorf("expected '%s' to follow", exact)
	}
	return nil
}
