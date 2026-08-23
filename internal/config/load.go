package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type loadError struct{ msg string }

func (e loadError) Error() string { return "subseapmp config: " + e.msg }
func errConfig(msg string) error  { return loadError{msg: msg} }

func Load() (Config, error) {
	cfg := Default()
	if v := os.Getenv("subseapmp_RACK_ID"); v != "" {
		cfg.StationID = strings.TrimSpace(v)
	}
	if v := os.Getenv("subseapmp_SLOT_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("subseapmp_SLOT_COUNT: %v", err))
		}
		cfg.SlotCount = n
	}
	if v := os.Getenv("subseapmp_DEFAULT_FLOW"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("subseapmp_DEFAULT_FLOW: %v", err))
		}
		cfg.DefaultFlowLPM = f
	}
	if v := os.Getenv("subseapmp_Pipeline_HOLD_MIN"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("subseapmp_Pipeline_HOLD_MIN: %v", err))
		}
		cfg.PressureHoldMinutes = n
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func IsConfigError(err error) bool {
	var le loadError
	return errors.As(err, &le)
}