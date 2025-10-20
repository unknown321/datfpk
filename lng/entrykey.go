package lng

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type entryKey uint32

func (k entryKey) MarshalJSON() ([]byte, error) {
	a := fmt.Sprintf(`"0x%x"`, k)
	return []byte(a), nil
}

func (k *entryKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if len(s) > 2 && s[0:2] == "0x" {
		s = s[2:]
	}

	val, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return err
	}

	*k = entryKey(val)

	return nil
}
