package crouton

import (
	"encoding/json"
	"time"
)

type Minutes time.Duration

func (min Minutes) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(min).Minutes())
}

func (m *Minutes) UnmarshalJSON(data []byte) error {
	var mins float64
	if err := json.Unmarshal(data, &mins); err != nil {
		return err
	}

	asM := Minutes(time.Duration(int64(mins * float64(time.Minute.Nanoseconds()))))
	m = &asM
	return nil
}
