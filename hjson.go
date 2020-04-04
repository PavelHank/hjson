package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	whitespace   = []byte{'\n', '\r', '\t', ' '}
)

func GetString(data []byte, target string) (string, error) {
	raw, err := Find(data, target)
	if err != nil {
		return "", err
	}
	if raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1:len(raw)-1]
	}
	return string(raw),nil
}

// GetAsNull would return nil with json null, otherwise error
func GetNull(data []byte, target string) error {
	raw, err := Find(data, target)
	if err != nil {
		return err
	}
	if string(raw) != `null` {
		return errors.New("expected null, but found:" + string(raw))
	}
	return nil
}

func GetInt(data []byte, target string) (int, error) {
	raw, err := Find(data, target)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(raw))
}

func GetBool(data []byte, target string) (bool, error) {
	raw, err := Find(data, target)
	if err != nil {
		return false, err
	}
	if bytes.ContainsAny(raw, `true`) {
		return true, nil
	}
	if bytes.ContainsAny(raw, `false`) {
		return false, nil
	}
	return false, errors.New("expected bool,but found:" + string(raw))
}

func GetFloat(data []byte, target string) (float64, error) {
	raw, err := Find(data, target)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(string(raw), 10)
}

func GetStringSlice(data []byte, target string) ([]string, error) {
	raw, err := Find(data, target)
	if err != nil {
		return nil, err
	}
	arr := bytes.Split(raw, []byte{','})
	var strSlice []string
	for _, v := range arr {
		strSlice = append(strSlice, string(bytes.Trim(v, "\" \n\t")))
	}
	return strSlice, nil
}

func trimWhitespace(data []byte) []byte {
	return bytes.Trim(data, " \n\t")
}

// Find each
func Find(data []byte, target string) ([]byte, error) {
	targets := strings.Split(target, ".")
	if len(targets) == 0 {
		return data, nil
	}
	tar := targets[0]
	data = trimWhitespace(data)
	if strings.HasPrefix(tar, `$`) && data[0] != '[' {
		return nil, errors.New("query an Array, but get invalid json string")
	}
	if data[0] != '{' && data[0] != '[' {
		return nil, errors.New("invalid jsonObject:" + string(data[0]) + " > " + target)
	}

	var rows [][]string
	var err error
	switch data[0] {
	case '{':
		rows, err = wrapObject(data)
	case '[':
		rows, err = wrapArray(data)
	default:
		return nil, errors.New("invalid json string, should start with `{` or `[`, but begin with:" + string(data[0]))
	}
	if err != nil {
		return nil, err
	}

	for _, v := range rows {
		if v[0] == tar || strings.Trim(v[0], `"`) == tar {
			if len(targets) == 1 {
				return []byte(v[1]), nil
			}
			return Find([]byte(v[1]), strings.Join(targets[1:], "."))
		}
	}
	return nil, errors.New("not found key:" + target)
}

func wrapArray(data []byte) ([][]string, error) {
	ln := len(data)
	i := 1
	var result [][]string
	for {
		if i >= ln {
			break
		}
		char := data[i]
		if inByteArr(whitespace, char) || char == ',' {
			i++
			continue
		}
		if char == ']' {
			break
		}
		validBeginChar := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '"', 't', 'f', '{', 'n'}
		if !inByteArr(validBeginChar, char) {
			return nil, errors.New("invalid json tokenize: " + string(char))
		}
		pair := make([]string, 2)
		pair[0] = fmt.Sprintf("$%d", len(result))
		lo, err := fetchElementValue(data[i:])
		if err != nil {
			return nil, err
		}
		pair[1] = string(data[i : i+lo])
		i += lo
		result = append(result, pair)
	}
	return result, nil
}

func fetchElementValue(data []byte) (int, error) {
	char := data[0]
	var lo int
	var err error
	switch char {
	case '"':
		lo, err = calculateStrLen(data)
	case '{':
		lo, err = calculateObjLen(data)
	case '[':
		lo, err = calculateArrayLen(data)
	case 'n':
		lo, err = calculateNullLen(data)
	case 'f', 't':
		lo, err = calculateBoolLen(data)
	case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		lo, err = calculateNumLen(data)
	default:
		return 0, errors.New("invalid json with unknown token: " + string(char))
	}
	return lo, err
}

func wrapObject(data []byte) ([][]string, error) {
	ln := len(data)
	i := 1
	var result [][]string
	for {
		if i >= ln {
			break
		}
		char := data[i]
		if inByteArr(whitespace, char) || char == ',' {
			i++
			continue
		}
		if char == '}' {
			break
		}

		if char != '"' {
			return nil, errors.New("invalid json tokenize: " + string(char))
		}

		pair := make([]string, 2)
		s, err := getCompleteString(data[i:])
		if err != nil {
			return nil, err
		}
		pair[0] = s
		i += len(s)
		hasComma := false
		for i < ln {
			char = data[i]
			if char == ':' {
				hasComma = true
			}
			if inByteArr(whitespace, char) || char == ':' {
				i++
			} else {
				break
			}
		}
		if !hasComma {
			return nil, errors.New("expected `:` but not found")
		}

		lo, err := fetchElementValue(data[i:])
		if err != nil {
			return nil, err
		}
		pair[1] = string(data[i : i+lo])
		i += lo
		result = append(result, pair)
	}
	return result, nil
}

func getCompleteString(data []byte) (string, error) {
	s := []byte{data[0]}
	sets := data[1:]
	endStr := false
	for i, c := range sets {
		s = append(s, c)
		if c == '"' && i > 0 && sets[i-1] != '\\' {
			endStr = true
			break
		}
	}
	if !endStr {
		return "", errors.New("invalid json string")
	}
	return string(s), nil
}

func calculateNumLen(data []byte) (int, error) {
	hasDot := false
	numSets := []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '.'}
	ln := 0
	for _, v := range data {
		if v == ',' || v == '}' || v == ']' {
			break
		}
		if !inByteArr(whitespace, v) && !inByteArr(numSets, v) {
			return 0, errors.New("invalid number serial")
		}
		if v == '.' {
			if hasDot {
				return 0, errors.New("invalid number serial:" + string(data))
			}
			hasDot = true
		}
		ln++
	}
	return ln, nil
}

func calculateNullLen(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("expected: `null`, but get:" + string(data))
	}
	if string(data[:4]) == "null" {
		return 4, nil
	}
	return 0, errors.New("invalid string, expect `null`,but get:" + string(data))
}

func calculateBoolLen(data []byte) (int, error) {
	if (data[0] == 'f' && len(data) < 5) || (data[0] == 't' && len(data) < 4) {
		return 0, errors.New("expected bool but get:" + string(data))
	}
	if data[0] == 'f' && string(data[:5]) == `false` {
		return 5, nil
	}
	if data[0] == 't' && string(data[:4]) == `true` {
		return 4, nil
	}
	return 0, errors.New("expected bool but get:" + string(data))
}

func calculateStrLen(data []byte) (int, error) {
	if data[0] != '"' {
		return 0, errors.New("invalid string of json:" + string(data))
	}
	ln := 1
	sets := data[1:]
	endStr := false
	for i, c := range sets {
		ln++
		if c == '"' && i > 0 && sets[i-1] != '\\' {
			endStr = true
			break
		}
	}
	if !endStr {
		return 0, errors.New("invalid json string")
	}
	return ln, nil
}

func calculateLen(data []byte, beginToken, endToken byte) (int, error) {
	var stack []byte
	for i, c := range data {
		if c == beginToken {
			stack = append(stack, c)
			continue
		}

		if c == endToken {
			if len(stack) == 0 {
				return 0, errors.New("invalid json on calculateObjectLength")
			}

			if len(stack) == 1 {
				return i + 1, nil
			}
			stack = stack[:len(stack)-1]
			continue
		}
	}
	return 0, errors.New("invalid json")
}

func calculateObjLen(data []byte) (int, error) {
	return calculateLen(data, '{', '}')
}

func calculateArrayLen(data []byte) (int, error) {
	return calculateLen(data, '[', ']')
}

func inByteArr(stack []byte, v byte) bool {
	return bytes.IndexByte(stack, v) > -1
}
