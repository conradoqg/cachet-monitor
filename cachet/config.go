package cachet

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type CachetMonitor struct {
	SystemName  string                   `json:"system_name" yaml:"system_name"`
	DateFormat  string                   `json:"date_format" yaml:"date_format"`
	API         CachetAPI                `json:"api"`
	Defconf     DefaultConfig            `json:"default_config" yaml:"default_config"`
	RawMonitors []map[string]interface{} `json:"monitors" yaml:"monitors"`
	RawConfigs  []map[string]interface{} `json:"configs" yaml:"configs"`
	Monitors    []MonitorInterface       `json:"-" yaml:"-"`
	Configs     []ConfigsInterface       `json:"-" yaml:"-"`
	Immediate   bool                     `json:"-" yaml:"-"`
}

var Mondef *DefaultConfig

// Validate configuration
func (cfg *CachetMonitor) Validate() bool {
	valid := true

	if len(cfg.SystemName) == 0 {
		// get hostname
		cfg.SystemName = getHostname()
	}

	if len(cfg.DateFormat) == 0 {
		cfg.DateFormat = DefaultTimeFormat
	}

	if len(cfg.API.Token) == 0 || len(cfg.API.URL) == 0 {
		logrus.Warnf("API URL or API Token missing.\nGet help at https://github.com/castawaylabs/cachet-monitor")
		valid = false
	}

	if len(cfg.Monitors) == 0 {
		logrus.Warnf("No monitors defined!\nSee help for example configuration")
		valid = false
	}

	Mondef = &cfg.Defconf

	for index, monitor := range cfg.Monitors {
		if errs := monitor.Validate(); len(errs) > 0 {
			logrus.Warnf("Monitor validation errors (index %d): %v", index, "\n - "+strings.Join(errs, "\n - "))
			valid = false
		}
	}

	return valid
}

// getHostname returns id of the current system
func getHostname() string {
	hostname, err := os.Hostname()
	if err == nil && len(hostname) > 0 {
		return hostname
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "unknown"
	}

	return addrs[0].String()
}

func getMs() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func GetMonitorType(t string) string {
	if len(t) == 0 {
		return "http"
	}

	return strings.ToLower(t)
}

func getTemplateData(monitor *AbstractMonitor) map[string]interface{} {
	return map[string]interface{}{
		"SystemName": monitor.config.SystemName,
		"API":        monitor.config.API,
		"Monitor":    monitor,
		"now":        time.Now().Format(monitor.config.DateFormat),
	}
}
