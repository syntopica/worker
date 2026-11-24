package api

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"unicode/utf8"
)

type body map[string]any

func decodeBody(raw []byte, emptyObject bool) (body, error) {
	if len(raw) == 0 && emptyObject {
		return body{}, nil
	}
	if !utf8.Valid(raw) || !json.Valid(raw) {
		return nil, badRequest()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded body
	if err := decoder.Decode(&decoded); err != nil || decoded == nil {
		return nil, badRequest()
	}
	return decoded, nil
}

func (b body) number(key string, fallback float64) (float64, error) {
	value, found := b[key]
	if !found {
		return fallback, nil
	}
	return numberValue(value)
}

func (b body) generation() (int64, error) {
	switch value := b["generation"].(type) {
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return integer, nil
		}
		return floatingGeneration(value)
	case string:
		return integerValue(strings.TrimSpace(value))
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, badRequest()
	}
}

func floatingGeneration(value json.Number) (int64, error) {
	number, err := value.Float64()
	if err != nil || number >= -float64(math.MinInt64) || number < float64(math.MinInt64) {
		return 0, badRequest()
	}
	return int64(number), nil
}
