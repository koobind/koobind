package common

import (
	"encoding/json"
	"errors"
	"time"
)


/*
A wrapper for Human readable JSON marshalling.
Work also with Encode/Decode

From: https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations

NB: I prefer this solution to

      type Duration time.Duration

as, for example, Add(MyDuration.Duration) is shorter than Add(time.Duration(MyDuration))
*/

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func ParseDuration(d string) (Duration, error) {
	duration, err := time.ParseDuration(d)
	if err == nil {
		return Duration{duration}, nil
	} else {
		return Duration{}, err
	}
}

func ParseDurationOrPanic(d string) Duration {
	duration, err := ParseDuration(d)
	if err != nil {
		panic(err)
	}
	return duration
}
