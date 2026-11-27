package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const maxSafeInteger = 9007199254740991

var uuid = regexp.MustCompile(`(?i)^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$`)

func ValidUUID(value string) bool {
	return uuid.MatchString(value)
}

func ValidText(value string, max int) bool {
	return strings.TrimSpace(value) != "" && utf8.RuneCountInString(value) <= max
}

func validDateTime(value string) bool {
	_, err := time.Parse(time.RFC3339Nano, value)
	return err == nil
}

// DecodeRequest checks exact field names and nulls before decoding to the transport type.
// Go's default decoder otherwise accepts case-insensitive names and null for scalars.
func DecodeRequest(reader io.Reader, target any, required, allowed string, minFields int) error {
	decoder := json.NewDecoder(reader)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil || decoder.Decode(new(any)) != io.EOF {
		return errors.New("se requiere exactamente un objeto JSON")
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil ||
		len(fields) < minFields || !validBody(fields, required, allowed) {
		return errors.New("campos ausentes, no admitidos o con formato incorrecto")
	}
	// Preserve JSON Schema's integer semantics for values such as 1.0.
	normalized, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	typed := json.NewDecoder(bytes.NewReader(normalized))
	typed.DisallowUnknownFields()
	return typed.Decode(target)
}

// validBody checks only the public JSON shape, never availability or business transitions.
func validBody(body map[string]any, required, allowed string) bool {
	for _, key := range strings.Fields(required) {
		if _, ok := body[key]; !ok {
			return false
		}
	}
	for key, value := range body {
		if value == nil || !slices.Contains(strings.Fields(allowed), key) {
			return false
		}
		switch key {
		case "expectedVersion":
			n, ok := value.(float64)
			if !ok || n < 1 || n > maxSafeInteger || n != float64(int64(n)) {
				return false
			}
		case "enabled":
			if _, ok := value.(bool); !ok {
				return false
			}
		case "startAt", "endAt":
			text, ok := value.(string)
			if !ok || !validDateTime(text) {
				return false
			}
		case "attributes":
			obj, ok := value.(map[string]any)
			if !ok || len(obj) != 0 {
				return false
			}
		case "policy":
			p, ok := value.(map[string]any)
			if !ok || len(p) != 4 {
				return false
			}
			for _, k := range []string{"cancellationNoticeSeconds", "minDurationSeconds", "maxDurationSeconds", "slotDurationSeconds"} {
				n, ok := p[k].(float64)
				if !ok || n != float64(int64(n)) || n < 0 || n > maxSafeInteger || (k != "cancellationNoticeSeconds" && n < 60) || (k == "slotDurationSeconds" && int64(n)%60 != 0) {
					return false
				}
			}
		case "externalReference":
			p, ok := value.(map[string]any)
			if !ok || !validBody(p, "system id", "system id") {
				return false
			}
		case "weeklySchedule":
			rules, ok := value.([]any)
			if !ok || len(rules) > 100 {
				return false
			}
			for _, rule := range rules {
				p, ok := rule.(map[string]any)
				if !ok || len(p) != 3 {
					return false
				}
				n, ok := p["dayOfWeek"].(float64)
				if !ok || n < 1 || n > 7 || n != float64(int(n)) {
					return false
				}
				for _, k := range []string{"startLocalTime", "endLocalTime"} {
					v, ok := p[k].(string)
					if !ok {
						return false
					}
					parsed, err := time.Parse("15:04", v)
					if err != nil || parsed.Format("15:04") != v {
						return false
					}
				}
			}
		case "resourceId":
			text, ok := value.(string)
			if !ok || !ValidUUID(text) {
				return false
			}
		default:
			text, ok := value.(string)
			limit := 128
			switch key {
			case "timeZone":
				limit = 64
			case "reason":
				limit = 500
			}
			if !ok || !ValidText(text, limit) {
				return false
			}
		}
	}
	return true
}
