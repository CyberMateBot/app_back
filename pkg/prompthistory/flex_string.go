package prompthistory

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// flexString accepts JSON string or number for telegramId fields.
type flexString string

func (s *flexString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*s = ""
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*s = flexString(asString)
		return nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(data, &asNumber); err == nil {
		*s = flexString(asNumber.String())
		return nil
	}

	var asFloat float64
	if err := json.Unmarshal(data, &asFloat); err == nil {
		*s = flexString(strconv.FormatInt(int64(asFloat), 10))
		return nil
	}

	return fmt.Errorf("flexString: unsupported value %s", string(data))
}

func (s flexString) String() string {
	return string(s)
}
