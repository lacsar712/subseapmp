package model

// PressureSetpoint defines allowable suction/discharge bands for subsea boosters.
type PressureSetpoint struct {
	MinSuctionBar   float64
	MaxDischargeBar float64
	MinDiffBar      float64
	MaxDiffBar      float64
}

func DefaultPressureSetpoint() PressureSetpoint {
	return PressureSetpoint{
		MinSuctionBar: 20, MaxDischargeBar: 250,
		MinDiffBar: 30, MaxDiffBar: 180,
	}
}

func (p PressureSetpoint) Valid() bool {
	return p.MinSuctionBar > 0 && p.MaxDischargeBar > p.MinSuctionBar &&
		p.MinDiffBar > 0 && p.MaxDiffBar >= p.MinDiffBar
}
