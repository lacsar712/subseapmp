package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("subseapmp: invalid identifier")
	ErrNotFound        = errors.New("subseapmp: entity not found")
	ErrConflict        = errors.New("subseapmp: state conflict")
	ErrInterlock       = errors.New("subseapmp: interlock denied")
	ErrPressureHold     = errors.New("subseapmp: Pipeline hold active")
	ErrFlowSetpoint    = errors.New("subseapmp: flow setpoint violation")
	ErrBooster      = errors.New("subseapmp: booster fault")
	ErrScheduleEmpty      = errors.New("subseapmp: schedule empty")
	ErrLinePressureHigh   = errors.New("subseapmp: line pressure high")
	ErrManifoldLeak       = errors.New("subseapmp: manifold leak")
	ErrInertHold          = errors.New("subseapmp: inert hold active")
	ErrContextCanceled    = errors.New("subseapmp: operation canceled")
)

type DomainError struct {
	Op   string
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("subseapmp %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("subseapmp %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }