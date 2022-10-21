package common

import "time"

var ISO8601Formats = []string{"2006-01-02T15:04:05.0000000Z", "2006-01-02T15:04:05Z", "2006-01-02T15:04Z", "2006-01-02"}

func TryParseISO8601(input string) *time.Time {
	for _, v := range ISO8601Formats {
		t, err := time.Parse(v, input)
		if err != nil {
			continue
		}

		return &t
	}

	return nil
}
