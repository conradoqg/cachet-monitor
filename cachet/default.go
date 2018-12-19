package cachet

import (
	"time"
)

const DefaultInterval = time.Second * 60
const DefaultTimeout = time.Second
const DefaultTimeFormat = "15:04:05 Jan 2 MST"
const DefaultHistorySize = 10
const DefaultTholdCrit = 60
const DefaultTholdPart = 40
const DefaultExpStsCode = 200
const DefContentType = "text/plain"

type DefaultConfig struct {
	DefInterval           int `json:"def_interval" yaml:"def_interval"`
	DefTimeout            int `json:"def_timeout" yaml:"def_timeout"`
	DefHistorySize        int `json:"def_history_size" yaml:"def_history_size"`
	DefThresholdCritical  int `json:"def_threshold_critical" yaml:"def_threshold_critical"`
	DefThresholdPartial   int `json:"def_threshold_partial" yaml:"def_threshold_partial"`
	DefExpectedStatusCode int `json:"def_expected_status_code" yaml:"def_expected_status_code"`

	DefWebhook struct {
		DefOnCritical struct {
			DefContentType string `yaml:"def_on_critical_content_type"`
			DefURL         string `yaml:"def_on_critical_url"`

			DefInvestigating struct {
				DefMessage string `yaml:"def_message"`
			} `yaml:"def_investigating"`
		} `yaml:"def_on_critical"`

		DefOnPartial struct {
			DefContentType string `yaml:"def_on_partial_content_type"`
			DefURL         string `yaml:"def_on_partial_url"`

			DefInvestigating struct {
				DefMessage string `yaml:"def_message"`
			} `yaml:"def_investigating"`
		} `yaml:"def_on_partial"`
	} `yaml:"def_webhook"`

	DefTemplate struct {
		DefInvestigating struct {
			DefSubJect string `yaml:"def_subject"`
			DefMessage string `yaml:"def_message"`
		} `yaml:"def_investigating"`

		DefFixed struct {
			DefSubJect string `yaml:"def_subject"`
			DefMessage string `yaml:"def_message"`
		} `yaml:"def_fixed"`
	} `yaml:"def_template"`
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

func (Def *DefaultConfig) GetDefTholdCritical() int {
	if Def.DefThresholdCritical <= 0 {
		return DefaultTholdCrit
	} else {
		return Def.DefThresholdCritical
	}

}

func (Def *DefaultConfig) GetDefTholdPartial() int {
	if Def.DefThresholdPartial <= 0 {
		return DefaultTholdPart
	} else {
		return Def.DefThresholdPartial
	}
}

func (Def *DefaultConfig) GetExpStsCode() int {
	if Def.DefExpectedStatusCode <= 0 {
		return DefaultExpStsCode
	} else {
		return Def.DefExpectedStatusCode
	}
}

func (Def *DefaultConfig) GetWCritContent() string {
	if len(Def.DefWebhook.DefOnCritical.DefContentType) == 0 {
		return DefContentType
	} else {
		return Def.DefWebhook.DefOnCritical.DefContentType
	}
}

func (Def *DefaultConfig) GetWCritUrl() string {
	return Def.DefWebhook.DefOnCritical.DefURL
}

func (Def *DefaultConfig) GetWCritMessage() string {
	return Def.DefWebhook.DefOnCritical.DefInvestigating.DefMessage
}

func (Def *DefaultConfig) GetWPartContent() string {
	if len(Def.DefWebhook.DefOnPartial.DefContentType) == 0 {
		return DefContentType
	} else {
		return Def.DefWebhook.DefOnPartial.DefContentType
	}
}

func (Def *DefaultConfig) GetWPartUrl() string {
	return Def.DefWebhook.DefOnPartial.DefURL
}

func (Def *DefaultConfig) GetWPartMessage() string {
	return Def.DefWebhook.DefOnCritical.DefInvestigating.DefMessage
}

func (Def *DefaultConfig) GetTempInvSub(s string) string {
	if len(Def.DefTemplate.DefInvestigating.DefSubJect) == 0 {
		return s
	} else {

		return Def.DefTemplate.DefInvestigating.DefSubJect
	}
}

func (Def *DefaultConfig) GetTempInvMes(s string) string {
	if len(Def.DefTemplate.DefInvestigating.DefMessage) == 0 {
		return s
	} else {
		return Def.DefTemplate.DefInvestigating.DefMessage
	}
}

func (Def *DefaultConfig) GetTempFixSub(s string) string {

	if len(Def.DefTemplate.DefFixed.DefSubJect) == 0 {
		return s
	} else {
		return Def.DefTemplate.DefFixed.DefSubJect
	}
}

func (Def *DefaultConfig) GetTempFixMes(s string) string {

	if len(Def.DefTemplate.DefFixed.DefMessage) == 0 {
		return s
	} else {
		return Def.DefTemplate.DefFixed.DefMessage
	}
}
