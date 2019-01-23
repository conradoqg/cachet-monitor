package cachet

import (
	"bytes"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type MonitorInterface interface {
	ClockStart(*CachetMonitor, MonitorInterface, *sync.WaitGroup)
	ClockStop()
	tick(MonitorInterface)
	test(l *logrus.Entry) bool

	Init(*CachetMonitor) bool
	Validate() []string
	GetMonitor() *AbstractMonitor
	Describe() []string
	//SetCron() string
}

// AbstractMonitor data model
type AbstractMonitor struct {
	Name    string
	Target  string
	Enabled bool

	// (default)http / dns
	Type   string
	Strict bool

	Interval time.Duration
	Timeout  time.Duration
	Resync   int

	MetricID    int `mapstructure:"metric_id"`
	ComponentID int `mapstructure:"component_id"`
	ConfigID    int `mapstructure:"config_id"`

	// Metric stuff
	Metrics struct {
		ResponseTime  []int `mapstructure:"response_time"`
		Availability  []int `mapstructure:"availability"`
		IncidentCount []int `mapstructure:"incident_count"`
	}

	ShellHook struct {
		// ShellHook stuff
		OnSuccess string `mapstructure:"on_success"`
		OnFailure string `mapstructure:"on_failure"`
	}

	Webhook struct {
		OnCritical struct {
			Investigating MessageTemplate
			ContentType   string `mapstructure:"content_type"`
			URL           string `mapstructure:"url"`
		} `mapstructure:"on_critical"`
		OnPartial struct {
			Investigating MessageTemplate
			ContentType   string `mapstructure:"content_type"`
			URL           string `mapstructure:"url"`
		} `mapstructure:"on_partial"`
	}

	// Templating stuff
	Template struct {
		Investigating MessageTemplate
		Fixed         MessageTemplate
	}

	// Threshold = percentage / number of down incidents
	HistorySize int `mapstructure:"history_size"`

	Threshold      int
	ThresholdCount int `mapstructure:"threshold_count"`

	CriticalThreshold      int `mapstructure:"threshold_critical"`
	CriticalThresholdCount int `mapstructure:"threshold_critical_count"`

	PartialThreshold      int `mapstructure:"threshold_partial"`
	PartialThresholdCount int `mapstructure:"threshold_partial_count"`

	// lag / average(lagHistory) * 100 = percentage above average lag
	// PerformanceThreshold sets the % limit above which this monitor will trigger degraded-performance
	// PerformanceThreshold float32

	resyncMod     int
	currentStatus int
	history       []bool
	// lagHistory     []float32
	lastFailReason string
	incident       *Incident
	config         *CachetMonitor

	// Closed when mon.Stop() is called
	stopC chan bool
}

func (mon *AbstractMonitor) Validate() []string {
	errs := []string{}

	if len(mon.Name) == 0 {
		errs = append(errs, "Name is required")
	}

	if mon.Interval < time.Second {
		mon.Interval = Mondef.GetDefInterval()

	}

	if mon.Timeout < time.Second {
		mon.Timeout = Mondef.GetDefTimeOut()
	}

	if mon.Timeout > mon.Interval {
		errs = append(errs, "Timeout greater than interval")

	}

	if mon.ComponentID == 0 && mon.MetricID == 0 {
		errs = append(errs, "component_id & metric_id are unset")
	}

	if mon.HistorySize <= 0 {
		mon.HistorySize = Mondef.GetDefHistorySize()
	}

	if mon.Threshold <= 0 {
		mon.Threshold = 0
	}

	if mon.CriticalThreshold <= 0 {
		mon.CriticalThreshold = Mondef.GetDefTholdCritical()
	}

	if mon.PartialThreshold <= 0 {
		mon.PartialThreshold = Mondef.GetDefTholdPartial()
	}

	if mon.Threshold == 0 && mon.CriticalThreshold == 0 && mon.PartialThreshold == 0 && mon.ThresholdCount == 0 && mon.CriticalThresholdCount == 0 && mon.PartialThresholdCount == 0 {
		mon.Threshold = 100
	}

	if len(mon.Webhook.OnCritical.ContentType) == 0 {
		mon.Webhook.OnCritical.ContentType = Mondef.GetWCritContent()
	}

	if len(mon.Webhook.OnCritical.URL) == 0 {
		mon.Webhook.OnCritical.URL = Mondef.GetWCritUrl()
	}

	if len(mon.Webhook.OnCritical.Investigating.Message) == 0 {
		mon.Webhook.OnCritical.Investigating.Message = Mondef.GetWCritMessage()
	}

	if len(mon.Webhook.OnPartial.ContentType) == 0 {
		mon.Webhook.OnPartial.ContentType = Mondef.GetWPartContent()
	}

	if len(mon.Webhook.OnPartial.URL) == 0 {
		mon.Webhook.OnPartial.URL = Mondef.GetWPartUrl()
	}

	if len(mon.Webhook.OnPartial.Investigating.Message) == 0 {
		mon.Webhook.OnPartial.Investigating.Message = Mondef.GetWPartMessage()
	}

	if mon.Template.Investigating.Subject == defaultHTTPInvestigatingTpl.Subject {
		mon.Template.Investigating.Subject = Mondef.GetTempInvSub(defaultHTTPInvestigatingTpl.Subject)
	}

	if mon.Template.Investigating.Message == defaultHTTPInvestigatingTpl.Message {
		mon.Template.Investigating.Message = Mondef.GetTempInvMes(defaultHTTPInvestigatingTpl.Message)
	}

	if mon.Template.Fixed.Subject == defaultHTTPFixedTpl.Subject {
		mon.Template.Fixed.Subject = Mondef.GetTempFixSub(defaultHTTPFixedTpl.Subject)
	}

	if mon.Template.Fixed.Message == defaultHTTPFixedTpl.Message {
		mon.Template.Fixed.Message = Mondef.GetTempFixMes(defaultHTTPFixedTpl.Message)
	}

	if err := mon.Template.Fixed.Compile(); err != nil {
		errs = append(errs, "Could not compile \"fixed\" template: "+err.Error())
	}
	if err := mon.Template.Investigating.Compile(); err != nil {
		errs = append(errs, "Could not compile \"investigating\" template: "+err.Error())
	}
	if err := mon.Webhook.OnPartial.Investigating.Compile(); err != nil {
		errs = append(errs, "Could not compile \"investigating\" template: "+err.Error())
	}
	if err := mon.Webhook.OnCritical.Investigating.Compile(); err != nil {
		errs = append(errs, "Could not compile \"investigating\" template: "+err.Error())
	}

	return errs
}
func (mon *AbstractMonitor) GetMonitor() *AbstractMonitor {
	return mon
}
func (mon *AbstractMonitor) Describe() []string {
	features := []string{"Type: " + mon.Type}

	if len(mon.Name) > 0 {
		features = append(features, "Name: "+mon.Name)
	}
	if len(mon.Target) > 0 {
		features = append(features, "Target: "+mon.Target)
	} else {
		features = append(features, "Target: <mock>")
	}
	features = append(features, "Availability count metrics: "+strconv.Itoa(len(mon.Metrics.Availability)))
	features = append(features, "Incident count metrics: "+strconv.Itoa(len(mon.Metrics.IncidentCount)))
	features = append(features, "Response time metrics: "+strconv.Itoa(len(mon.Metrics.ResponseTime)))
	if mon.Resync > 0 {
		features = append(features, "Resyncs cycle: "+strconv.Itoa(mon.Resync))
	}
	if len(mon.ShellHook.OnSuccess) > 0 {
		features = append(features, "Has a 'on_success' shellhook")
	}
	if len(mon.ShellHook.OnFailure) > 0 {
		features = append(features, "Has a 'on_failure' shellhook")
	}

	return features
}

func (mon *AbstractMonitor) ReloadCachetData() {
	compInfo := mon.config.API.GetComponentData(mon.ComponentID)

	logrus.Infof("Current CachetHQ ID: %d", compInfo.ID)
	logrus.Infof("Current CachetHQ name: %s", compInfo.Name)
	logrus.Infof("Current CachetHQ state: %t", compInfo.Enabled)
	logrus.Infof("Current CachetHQ status: %d", compInfo.Status)
	if mon.ThresholdCount > 0 || mon.Threshold > 0 {
		if mon.ThresholdCount > 0 {
			logrus.Infof("Threshold (count): %d", mon.ThresholdCount)
		} else {
			logrus.Infof("Threshold (percent): %d", mon.Threshold)
		}
	} else {
		if mon.CriticalThresholdCount > 0 || mon.CriticalThreshold > 0 {
			if mon.CriticalThresholdCount > 0 {
				logrus.Infof("Critical threshold (count): %d", mon.CriticalThresholdCount)
			} else {
				logrus.Infof("Critical threshold (percent): %d", mon.CriticalThreshold)
			}
		}
		if mon.PartialThresholdCount > 0 || mon.PartialThreshold > 0 {
			if mon.PartialThresholdCount > 0 {
				logrus.Infof("Partial threshold (count): %d", mon.PartialThresholdCount)
			} else {
				logrus.Infof("Partial threshold (percent): %d", mon.PartialThreshold)
			}
		}
	}
	//logrus.Info("DADO: %d", mon.HistorySize)

	mon.currentStatus = compInfo.Status
	mon.Enabled = compInfo.Enabled

	mon.incident, _ = compInfo.LoadCurrentIncident(mon.config)

	if mon.incident != nil {
		logrus.Infof("Current incident ID: %v", mon.incident.ID)
	} else {
		logrus.Infof("No current incident")
	}
}

func (mon *AbstractMonitor) Init(cfg *CachetMonitor) bool {
	mon.config = cfg

	IsValid := true
	// logrus.Infof("ALTEREI AS PARADAS, %d",mon.HistorySize)
	mon.ReloadCachetData()

	// if mon.HistorySize == 0 {
	// 	mon.HistorySize = 88
	// }

	if mon.ComponentID == 0 {
		logrus.Infof("ComponentID couldn't be retreived")
		IsValid = false
	}

	mon.history = append(mon.history, mon.isUp())

	return IsValid
}

func (mon *AbstractMonitor) triggerShellHook(l *logrus.Entry, hooktype string, hook string, data string) {
	if len(hook) == 0 {
		return
	}
	l.Infof("Sending '%s' shellhook", hooktype)
	l.Debugf("Data: %s", data)

	out, err := exec.Command(hook, mon.Name, strconv.Itoa(mon.ComponentID), mon.Target, hooktype, data).Output()
	if err != nil {
		l.Warnf("Error when processing shellhook '%s': %s", hooktype, err)
		l.Warnf("Command output: %s", out)
	}
}

func (mon *AbstractMonitor) ClockStart(cfg *CachetMonitor, iface MonitorInterface, wg *sync.WaitGroup) {
	wg.Add(1)

	mon.stopC = make(chan bool)

	if cfg.Immediate {
		mon.tick(iface)
	}

	ticker := time.NewTicker(mon.Interval * time.Second)
	for {
		select {
		case <-ticker.C:
			mon.tick(iface)
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (mon *AbstractMonitor) ClockStop() {
	select {
	case <-mon.stopC:
		return
	default:
		close(mon.stopC)
	}
}

func (mon *AbstractMonitor) isUp() bool {
	return (mon.currentStatus == 1)
}

func (mon *AbstractMonitor) isPartial() bool {
	return (mon.currentStatus == 3)
}

func (mon *AbstractMonitor) isCritical() bool {
	return (mon.currentStatus == 4)
}

func (mon *AbstractMonitor) test(l *logrus.Entry) bool { return false }

func (mon *AbstractMonitor) tick(iface MonitorInterface) {
	l := logrus.WithFields(logrus.Fields{"monitor": mon.Name})

	if !mon.Enabled {
		l.Printf("monitor is disabled")
		return
	}

	reqStart := getMs()
	isUp := true
	isUp = iface.test(l)
	lag := getMs() - reqStart

	if !isUp {
		lag = 0
	}

	if len(mon.history) == mon.HistorySize-1 {
		l.Debugf("monitor %v is now fully operational", mon.Name)
	}

	if len(mon.history) >= mon.HistorySize {
		mon.history = mon.history[len(mon.history)-(mon.HistorySize-1):]
	}
	mon.history = append(mon.history, isUp)

	mon.AnalyseData(l)

	// Will trigger shellhook 'on_failure' as this isn't done in implementations
	if !isUp {
		mon.triggerShellHook(l, "on_failure", mon.ShellHook.OnFailure, "")
	}

	// report lag
	if mon.MetricID > 0 {
		go mon.config.API.SendMetric(l, mon.MetricID, lag)
	}
	go mon.config.API.SendMetrics(l, "response time", mon.Metrics.ResponseTime, lag)

	if mon.Resync > 0 {
		mon.resyncMod = (mon.resyncMod + 1) % mon.Resync
		if mon.resyncMod == 0 {
			l.Debugf("Reloading component's data")
			mon.ReloadCachetData()
		} else {
			l.Debugf("Resync progressbar: %d/%d", mon.resyncMod, mon.Resync)
		}
	}
}

// TODO: test
// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (mon *AbstractMonitor) AnalyseData(l *logrus.Entry) {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range mon.history {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(mon.history))) * 100
	if numDown == 0 {
		l.Printf("monitor is fully up")
		go mon.config.API.SendMetrics(l, "availability", mon.Metrics.Availability, 1)
	}

	if len(mon.history) != mon.HistorySize {
		// not yet saturated
		l.Debugf("Component's history has not been yet saturated (stack: %d/%d)", len(mon.history), mon.HistorySize)
		return
	}

	triggered := false
	criticalTriggered := false
	partialTriggered := false

	if numDown > 0 {
		if mon.ThresholdCount > 0 || mon.Threshold > 0 {
			if mon.ThresholdCount > 0 {
				triggered = (numDown >= mon.ThresholdCount)
				l.Printf("monitor down (down count=%d, threshold=%d)", numDown, mon.Threshold)
			} else {
				triggered = (int(t) > mon.Threshold)
				l.Printf("monitor down (down percentage=%.2f%%, threshold=%d%%)", t, mon.Threshold)
			}
		} else {
			if mon.CriticalThresholdCount > 0 || mon.CriticalThreshold > 0 {
				if mon.CriticalThresholdCount > 0 {
					criticalTriggered = (numDown >= mon.CriticalThresholdCount)
				} else {
					criticalTriggered = (int(t) > mon.CriticalThreshold)
				}
			}
			if !criticalTriggered {
				if mon.PartialThresholdCount > 0 || mon.PartialThreshold > 0 {
					partialTriggered = (mon.PartialThresholdCount > 0 && numDown >= mon.PartialThresholdCount) || (mon.PartialThreshold > 0 && int(t) > mon.PartialThreshold)
				}
			}
			if mon.CriticalThresholdCount > 0 || mon.PartialThresholdCount > 0 {
				l.Printf("monitor down (down count=%d, partial threshold=%d, critical threshold=%d)", numDown, mon.PartialThresholdCount, mon.CriticalThresholdCount)
			}
			if mon.CriticalThreshold > 0 || mon.PartialThreshold > 0 {
				l.Printf("monitor down (down percentage=%.2f%%, partial threshold=%d%%, critical threshold=%d%%)", t, mon.PartialThreshold, mon.CriticalThreshold)
			}
		}
		l.Debugf("Down count: %d, history: %d, percentage: %.2f%%", numDown, len(mon.history), t)
		l.Debugf("Is triggered: %t", triggered)
		l.Debugf("Is critically Triggered: %t", criticalTriggered)
		l.Debugf("Is partially Triggered: %t", partialTriggered)
		l.Debugf("Monitor's current incident: %v", mon.incident)

		if triggered || criticalTriggered || partialTriggered {
			// Process metric
			go mon.config.API.SendMetrics(l, "incident count", mon.Metrics.IncidentCount, 1)

			if mon.incident == nil {
				// create incident
				mon.currentStatus = 2
				tplData := getTemplateData(mon)
				tplData["FailReason"] = mon.lastFailReason

				subject, message := mon.Template.Investigating.Exec(tplData)
				incidentForceComponentStatus := 4
				if partialTriggered {
					incidentForceComponentStatus = 3
				}
				mon.incident = &Incident{
					Name:            subject,
					ComponentID:     mon.ComponentID,
					Message:         message,
					Notify:          true,
					ComponentStatus: incidentForceComponentStatus,
				}

				// is down, create an incident
				l.Warnf("creating incident. Monitor is down: %v", mon.lastFailReason)
				// set investigating status
				mon.incident.SetInvestigating()
				// create incident
				if err := mon.incident.Send(mon.config); err != nil {
					l.Printf("Error sending incident: %v", err)
				}

				// call webhook

				if partialTriggered && len(mon.Webhook.OnPartial.URL) > 0 {
					l.Debugf("Calling OnPartial webhook")

					message := mon.Webhook.OnPartial.Investigating.ExecMessage(tplData)

					url := mon.Webhook.OnPartial.URL

					req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(message)))
					req.Header.Set("Content-Type", mon.Webhook.OnPartial.ContentType)

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						l.Printf("Error calling webhook: %v", err)
					}
					defer resp.Body.Close()
				}

				if criticalTriggered && len(mon.Webhook.OnCritical.URL) > 0 {
					l.Debugf("Calling OnCritical webhook")

					message := mon.Webhook.OnCritical.Investigating.ExecMessage(tplData)

					url := mon.Webhook.OnCritical.URL

					req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(message)))
					req.Header.Set("Content-Type", mon.Webhook.OnCritical.ContentType)

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						l.Printf("Error calling webhook: %v", err)
					}
					defer resp.Body.Close()
				}
			}
			if triggered || criticalTriggered {
				if !mon.isCritical() {
					mon.config.API.SetComponentStatus(mon, 4)
				}
			}
			if partialTriggered {
				if !mon.isPartial() {
					mon.config.API.SetComponentStatus(mon, 3)
				}
			}
			return
		}
	}

	// we are up to normal

	// global status seems incorrect though we couldn't fid any prior incident
	if !mon.isUp() && mon.incident == nil {
		l.Info("Reseting component's status")
		mon.lastFailReason = ""
		mon.incident = nil
		mon.config.API.SetComponentStatus(mon, 1)
		return
	}

	if mon.incident == nil {
		return
	}

	// was down, created an incident, its now ok, make it resolved.
	l.Infof("Resolving incident %d", mon.incident.ID)

	// resolve incident
	tplData := getTemplateData(mon)
	tplData["incident"] = mon.incident

	subject, message := mon.Template.Fixed.Exec(tplData)
	mon.incident.Name = subject
	mon.incident.Message = message
	mon.incident.SetFixed()
	if err := mon.incident.Send(mon.config); err != nil {
		l.Warnf("Error updating sending incident: %v", err)
	}

	mon.lastFailReason = ""
	mon.incident = nil
	mon.currentStatus = 1
}
