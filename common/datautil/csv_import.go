package datautil

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gorse-io/gorse/storage/data"
)

type CSVMapper struct {
	FeedbackFields map[string]string
	ItemFields     map[string]string
	UserFields     map[string]string
}

var DefaultCSVMapper = CSVMapper{
	FeedbackFields: map[string]string{
		"user_id":       "UserId",
		"userid":        "UserId",
		"user":          "UserId",
		"item_id":       "ItemId",
		"itemid":        "ItemId",
		"item":          "ItemId",
		"feedback_type": "FeedbackType",
		"feedbacktype":  "FeedbackType",
		"type":          "FeedbackType",
		"timestamp":     "Timestamp",
		"time":          "Timestamp",
		"date":          "Timestamp",
		"value":         "Value",
	},
	ItemFields: map[string]string{
		"item_id":   "ItemId",
		"itemid":    "ItemId",
		"item":      "ItemId",
		"is_hidden": "IsHidden",
		"ishidden":  "IsHidden",
		"hidden":    "IsHidden",
		"categories": "Categories",
		"category":  "Categories",
		"timestamp": "Timestamp",
		"time":      "Timestamp",
		"date":      "Timestamp",
		"labels":    "Labels",
		"label":     "Labels",
		"comment":   "Comment",
	},
	UserFields: map[string]string{
		"user_id": "UserId",
		"userid":  "UserId",
		"user":    "UserId",
		"labels":  "Labels",
		"label":   "Labels",
		"comment": "Comment",
	},
}

func AutoDetectMapping(headers []string, entityType string) (map[int]string, error) {
	var fields map[string]string
	switch strings.ToLower(entityType) {
	case "feedback":
		fields = DefaultCSVMapper.FeedbackFields
	case "item":
		fields = DefaultCSVMapper.ItemFields
	case "user":
		fields = DefaultCSVMapper.UserFields
	default:
		return nil, fmt.Errorf("csv import: unknown entity type %q", entityType)
	}

	mapping := make(map[int]string, len(headers))
	var unrecognized []string
	for i, header := range headers {
		normalized := strings.TrimSpace(strings.ToLower(header))
		if field, ok := fields[normalized]; ok {
			mapping[i] = field
		} else {
			unrecognized = append(unrecognized, header)
		}
	}
	if len(unrecognized) > 0 {
		return mapping, fmt.Errorf("csv import: unrecognized column names: %s", strings.Join(unrecognized, ", "))
	}
	return mapping, nil
}

func ImportFeedbackFromCSV(reader io.Reader) ([]data.Feedback, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv import: failed to read headers: %w", err)
	}
	mapping, err := AutoDetectMapping(headers, "feedback")
	if err != nil {
		return nil, err
	}

	var feedbacks []data.Feedback
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv import: failed to read record: %w", err)
		}

		var fb data.Feedback
		for i, field := range mapping {
			if i >= len(record) {
				continue
			}
			value := strings.TrimSpace(record[i])
			if value == "" {
				continue
			}
			switch field {
			case "UserId":
				fb.UserId = value
			case "ItemId":
				fb.ItemId = value
			case "FeedbackType":
				fb.FeedbackType = value
			case "Timestamp":
				fb.Timestamp, err = dateparse.ParseAny(value)
				if err != nil {
					return nil, fmt.Errorf("csv import: failed to parse timestamp %q: %w", value, err)
				}
			case "Value":
				fb.Value, err = strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, fmt.Errorf("csv import: failed to parse value %q: %w", value, err)
				}
			}
		}
		feedbacks = append(feedbacks, fb)
	}
	return feedbacks, nil
}

func ImportItemsFromCSV(reader io.Reader) ([]data.Item, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv import: failed to read headers: %w", err)
	}
	mapping, err := AutoDetectMapping(headers, "item")
	if err != nil {
		return nil, err
	}

	var items []data.Item
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv import: failed to read record: %w", err)
		}

		var item data.Item
		for i, field := range mapping {
			if i >= len(record) {
				continue
			}
			value := strings.TrimSpace(record[i])
			if value == "" {
				continue
			}
			switch field {
			case "ItemId":
				item.ItemId = value
			case "IsHidden":
				item.IsHidden, err = strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("csv import: failed to parse is_hidden %q: %w", value, err)
				}
			case "Categories":
				item.Categories = strings.Split(value, ",")
				for j, cat := range item.Categories {
					item.Categories[j] = strings.TrimSpace(cat)
				}
			case "Timestamp":
				item.Timestamp, err = dateparse.ParseAny(value)
				if err != nil {
					return nil, fmt.Errorf("csv import: failed to parse timestamp %q: %w", value, err)
				}
			case "Labels":
				raw := strings.Split(value, ",")
				trimmed := make([]string, len(raw))
				for j, l := range raw {
					trimmed[j] = strings.TrimSpace(l)
				}
				item.Labels = trimmed
			case "Comment":
				item.Comment = value
			}
		}
		if item.Timestamp.IsZero() {
			item.Timestamp = time.Now()
		}
		items = append(items, item)
	}
	return items, nil
}

func ImportUsersFromCSV(reader io.Reader) ([]data.User, error) {
	csvReader := csv.NewReader(reader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv import: failed to read headers: %w", err)
	}
	mapping, err := AutoDetectMapping(headers, "user")
	if err != nil {
		return nil, err
	}

	var users []data.User
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv import: failed to read record: %w", err)
		}

		var user data.User
		for i, field := range mapping {
			if i >= len(record) {
				continue
			}
			value := strings.TrimSpace(record[i])
			if value == "" {
				continue
			}
			switch field {
			case "UserId":
				user.UserId = value
			case "Labels":
				raw := strings.Split(value, ",")
				trimmed := make([]string, len(raw))
				for j, l := range raw {
					trimmed[j] = strings.TrimSpace(l)
				}
				user.Labels = trimmed
			case "Comment":
				user.Comment = value
			}
		}
		users = append(users, user)
	}
	return users, nil
}
