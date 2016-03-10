package slack

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// A Timestamp is a special time.Time alias that parses Slack timestamps better.
type Timestamp struct {
	time time.Time
	uuid string
}

func (t Timestamp) String() string {
	if len(t.uuid) != 0 {
		return fmt.Sprintf("%d.%s", t.time.Unix(), t.uuid)
	}
	return fmt.Sprintf("%d", t.time.Unix())
}

// UnmarshalJSON implements json.Unmarshal for the Timestamp struct.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	strValue := string(data)
	strValue = strings.Replace(strValue, `"`, ``, -1)
	if strings.Contains(strValue, ".") {
		components := strings.Split(strValue, ".")
		if integerValue, integerErr := strconv.ParseInt(components[0], 10, 64); integerErr == nil {
			t.time = time.Unix(integerValue, 0)
			t.uuid = components[1]
		}
	}

	if integerValue, integerErr := strconv.ParseInt(strValue, 10, 64); integerErr == nil {
		t.time = time.Unix(integerValue, 0)
	}

	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// DateTime returns a regular golang time.Time for the Timestamp instance.
func (t Timestamp) Time() time.Time {
	return t.time
}

func (t Timestamp) UUID() string {
	return t.uuid
}
