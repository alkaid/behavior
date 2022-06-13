package util

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Duration struct {
	time.Duration
}

func (duration *Duration) UnmarshalJSON(b []byte) error {
	var unmarshalledJson interface{}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return errors.WithStack(err)
	}

	switch value := unmarshalledJson.(type) {
	case float64:
		duration.Duration = time.Duration(value)
	case string:
		if value == "" {
			duration.Duration = 0
			return nil
		}
		duration.Duration, err = time.ParseDuration(value)
		if err != nil {
			return errors.WithStack(err)
		}
	default:
		return errors.WithStack(fmt.Errorf("invalid duration: %#v", unmarshalledJson))
	}

	return nil
}
