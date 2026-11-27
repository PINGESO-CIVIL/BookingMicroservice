package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"bookingms/internal/model/demo"
)

func (s *Service) listResources(input Input) (any, *Fault) {
	filtered := []demo.Resource{}
	if input.Scenario != "empty" {
		for _, row := range s.fixtures.Resources() {
			if row.Enabled == input.Query.Enabled {
				filtered = append(filtered, row)
			}
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	return page(input, filtered)
}

func (s *Service) listReservations(admin bool, input Input) (any, *Fault) {
	q := input.Query
	filtered := []demo.Reservation{}
	if input.Scenario != "empty" {
		for _, row := range s.fixtures.Reservations(admin) {
			if (q.ResourceID == "" || row.ResourceID == q.ResourceID) &&
				(q.Status == "" || row.Status == q.Status) &&
				(q.From.IsZero() || overlaps(row.Interval, q.From, q.To)) {
				filtered = append(filtered, row)
			}
		}
	}
	return page(input, filtered)
}

type cursor struct {
	Offset int
	Query  string
}

func page[T demo.Resource | demo.Reservation](input Input, rows []T) (demo.Page[T], *Fault) {
	q := input.Query
	limit := q.Limit
	if input.Scenario == "paginated" {
		limit = 2
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(q.CursorScope)))
	offset := 0
	if q.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(q.Cursor)
		var c cursor
		if err != nil || json.Unmarshal(raw, &c) != nil || c.Query != hash || c.Offset < 0 || c.Offset > len(rows) {
			return demo.Page[T]{}, &Fault{Message: "Cursor inválido para esta consulta."}
		}
		offset = c.Offset
	}
	end := min(offset+limit, len(rows))
	var next *string
	if end < len(rows) {
		raw, _ := json.Marshal(cursor{Offset: end, Query: hash})
		value := base64.RawURLEncoding.EncodeToString(raw)
		next = &value
	}
	return demo.Page[T]{Items: rows[offset:end], NextCursor: next}, nil
}

func (s *Service) availability(input Input) (any, *Fault) {
	from, to := input.Query.From, input.Query.To
	if !from.Before(to) || to.Sub(from) > 31*24*time.Hour {
		return nil, &Fault{Message: "Consulta un rango válido de hasta 31 días."}
	}
	result := s.fixtures.Availability(input.ID)
	windows, slots := result.AvailableWindows, result.AvailableSlots
	result.From = from.UTC().Format(time.RFC3339Nano)
	result.To = to.UTC().Format(time.RFC3339Nano)
	result.AvailableWindows, result.AvailableSlots = []demo.Interval{}, []demo.Interval{}
	if input.Scenario != "empty" && input.Scenario != "disabled" {
		for _, window := range windows {
			start, end := intervalTimes(window)
			if !start.Before(to) || !end.After(from) {
				continue
			}
			if start.Before(from) {
				window.StartAt = result.From
			}
			if end.After(to) {
				window.EndAt = result.To
			}
			result.AvailableWindows = append(result.AvailableWindows, window)
		}
		for _, slot := range slots {
			start, end := intervalTimes(slot)
			if !start.Before(from) && !end.After(to) {
				result.AvailableSlots = append(result.AvailableSlots, slot)
			}
		}
	}
	if input.Scenario == "disabled" {
		result.Enabled = false
	}
	return result, nil
}

func intervalTimes(interval demo.Interval) (time.Time, time.Time) {
	start, _ := time.Parse(time.RFC3339Nano, interval.StartAt)
	end, _ := time.Parse(time.RFC3339Nano, interval.EndAt)
	return start, end
}

func overlaps(interval demo.Interval, from, to time.Time) bool {
	start, end := intervalTimes(interval)
	return start.Before(to) && end.After(from)
}
