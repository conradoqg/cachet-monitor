package cachet

import (
	"time"
)

const DefaultInterval = time.Second * 60
const DefaultTimeout = time.Second
const DefaultTimeFormat = "15:04:05 Jan 2 MST"
const DefaultHistorySize = 10
const DefaultthresholdCritical = 60
const DefaultthresholdPartial = 40
const DefaultExpectedStatusCode = 200

type DefaultConfig struct {
	DefInterval           int `json:"def_interval" yaml:"def_interval"`
	DefTimeout            int `json:"def_timeout" yaml:"def_timeout"`
	DefHistorySize        int `json:"def_history_size" yaml:"def_history_size"`
	DefThresholdCritical  int `json:"def_threshold_critical" yaml:"def_threshold_critical"`
	DefThresholdPartial   int `json:"def_threshold_partial" yaml:"def_threshold_partial"`
	DefExpectedStatusCode int `json:"def_expected_status_code" yaml:"def_expected_status_code"`
}

func (Def *DefaultConfig) GetDefInterval() time.Duration {
	if Def.DefInterval <= 0 {
		return DefaultInterval
	} else {
		return time.Second * time.Duration(Def.DefInterval)
	}

}

func (Def *DefaultConfig) GetDefTimeOut() time.Duration {
	if Def.DefTimeout <= 0 {
		return DefaultTimeout
	} else {
		return time.Second * time.Duration(Def.DefTimeout)
	}

}

func (Def *DefaultConfig) GetDefHistorySize() int {
	if Def.DefHistorySize <= 0 {
		return DefaultHistorySize
	} else {
		return Def.DefHistorySize
	}

}

func (Def *DefaultConfig) GetDefThresholdCritical() int {
	if Def.DefThresholdCritical <= 0 {
		return DefaultthresholdCritical
	} else {
		return Def.DefThresholdCritical
	}

}

func (Def *DefaultConfig) GetDefThresholdPartial() int {
	if Def.DefThresholdPartial <= 0 {
		return DefaultthresholdPartial
	} else {
		return Def.DefThresholdPartial
	}

}

func (Def *DefaultConfig) GetExpectedStatusCode() int {
	if Def.DefExpectedStatusCode <= 0 {
		return DefaultExpectedStatusCode
	} else {
		return Def.DefExpectedStatusCode
	}

}
