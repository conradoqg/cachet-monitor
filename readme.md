![screenshot](https://conradoqg.github.io/cachet-monitor/screenshot.png)

## Features

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag, availability, incident count to cachet graphs
- [x] HTTP Checks (body/status code)
- [x] DNS Checks
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if already in Partial Outage (works with distributed monitors)
- [x] Can be run on multiple servers and geo regions

## Example Configuration

**Note:** configuration can be in json or yaml format. [`example.config.json`](https://github.com/conradoqg/cachet-monitor/blob/master/example.config.json), [`example.config.yaml`](https://github.com/conradoqg/cachet-monitor/blob/master/example.config.yml) files.

```yaml
api:
  # cachet url
  url: https://demo.cachethq.io/api/v1
  # cachet api token
  token: 9yMHsdioQosnyVK4iCVR
  insecure: false
# https://golang.org/src/time/format.go#L57
date_format: 02/01/2006 15:04:05 MST
monitors:
  # http monitor example
  - name: google
    # test url
    target: https://google.com
    # strict certificate checking for https
    strict: true
    # HTTP method
    method: POST
    
    # set to update component (either component_id or metric_id/metrics are required)
    component_id: 1
    
    # set to post to cachet metrics [ response_time_metric_id, availability_metric_id, incident_count_metric_id ] or metric_id: response_time_metric_id  (graph)
    metrics:
        response_time: [ 4, 5 ]

    # custom templates (see readme for details)
    template:
      investigating:
        subject: "{{ .Monitor.Name }} - {{ .SystemName }}"
        message: "{{ .Monitor.Name }} check **failed** (server time: {{ .now }})\n\n{{ .FailReason }}"
      fixed:
        subject: "I HAVE BEEN FIXED"

    # launch script depending on event (failed or successful check)
    shellhook:    
      on_success: /fullpath/shellhook_onsuccess.sh
      on_failure: /fullpath/shellhook_onfailure.sh

    # webhook to be called when a partial occurs
    webhook:
      on_partial:
        url: "http://www.site.com/webhook"
        content_type: "text/plain"
        investigating:        
          message: "{{ .Monitor.Name }} check **failed** (server time: {{ .now }})\n\n{{ .FailReason }}"
      on_critical:
        url: "http://www.site.com/webhook"
        content_type: "text/plain"
        investigating:        
          message: "{{ .Monitor.Name }} check **failed** (server time: {{ .now }})\n\n{{ .FailReason }}"
    
    # seconds between checks
    interval: 1
    # seconds for timeout
    timeout: 1

    # resync component data every x check
    resync: 60

    # necessary ticks before saturation (before evaluating the downtime)
    history_size: 10

    # if % or (count: threshold_count) of downtime is over this threshold, open an incident
    threshold: 50

    # or if % or (count: threshold_critical_count/threshold_partial_count) of downtime is over partical/critical open an incident with the related incident level
    threshold_critical: 80
    threshold_partial: 20

    # custom HTTP headers
    headers:
      Authorization: Basic <hash>
    # expected status code (either status code or body must be supplied)
    expected_status_code: 200
    # regex to match body
    expected_body: "P.*NG"

  # dns monitor example
  - name: dns
    # fqdn
    target: matej.me.
    # question type (A/AAAA/CNAME/...)
    question: mx
    type: dns
    # set component_id/metric_id
    component_id: 2
    # poll every 1s
    interval: 1
    timeout: 1
    # custom DNS server (defaults to system)
    dns: 8.8.4.4:53
    answers:
      - exact: 10 aspmx2.googlemail.com.
      - exact: 1 aspmx.l.google.com.
      - exact: 10 aspmx3.googlemail.com.
```

## Installation

1. Download binary from [release page](https://github.com/conradoqg/cachet-monitor/releases)
2. Create a configuration
3. `cachet-monitor -c /etc/cachet-monitor.yaml`

pro tip: run in background using `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`

```
Usage:
  cachet-monitor (-c PATH | --config PATH)
  cachet-monitor (-c PATH | --config PATH) [--log=LOGPATH] [--name=NAME] [--immediate] [--config-test] [--log-level=LOGLEVEL]
  cachet-monitor -h | --help | --version

Arguments:
  PATH     path to config.json
  LOGLEVEL log level (debug, info, warn, error or fatal)
  LOGPATH  path to log output (defaults to STDOUT)
  NAME     name of this logger

Examples:
  cachet-monitor -c /root/cachet-monitor.json
  cachet-monitor -c /root/cachet-monitor.json --log=/var/log/cachet-monitor.log --name="development machine"

Options:
  -h --help                      Show this screen.
  -c PATH.json --config PATH     Path to configuration file
  [--log]		                 Sets log file
  [--log-level]		         Sets log level
  [--config-test]                Check configuration file
  [--version]                      Show version
  [--immediate]                    Tick immediately (by default waits for first defined interval)
  
Environment varaibles:
  CACHET_API      override API url from configuration
  CACHET_TOKEN    override API token from configuration
  CACHET_DEV      set to enable dev logging
```

## Init script

If your system is running systemd (like Debian, Ubuntu 16.04, Fedora or Archlinux) you can use the provided example file: [example.cachet-monitor.service](https://github.com/conradoqg/cachet-monitor/blob/master/example.cachet-monitor.service).

1. Simply put it in the right place with `cp example.cachet-monitor.service /etc/systemd/system/cachet-monitor.service`
2. Then do a `systemctl daemon-reload` in your terminal to update Systemd configuration
3. Finally you can start cachet-monitor on every startup with `systemctl enable cachet-monitor.service`! 👍

## Templates

This package makes use of [`text/template`](https://godoc.org/text/template). [Default HTTP template](https://github.com/conradoqg/cachet-monitor/blob/master/http.go#L14)

The following variables are available:

| Root objects  | Description                         |
| ------------- | -----------------	+| ------------- | ------------------------------------|
| `.SystemName` | system name	+| `.SystemName` | system name                         | 
| `.API`        | `api` object from configuration	+| `.API`        | `api` object from configuration     |
| `.Monitor`    | `monitor` object from configuration	+| `.Monitor`    | `monitor` object from configuration |
| `.now`        | formatted date string	+| `.now`        | formatted date string               |

| Monitor variables  |
| ------------------ |
| `.Name`            |
| `.Target`          |
| `.Type`            |
| `.Strict`          |
| `.MetricID`        |
| ...                |

All monitor variables are available from `monitor.go`

## Vision and goals

We made this tool because we felt the need to have our own monitoring software (leveraging on Cachet).
The idea is a stateless program which collects data and pushes it to a central cachet instance.

This gives us power to have an army of geographically distributed loggers and reveal issues in both latency & downtime on client websites.

## Package usage

When using `cachet-monitor` as a package in another program, you should follow what `cli/main.go` does. It is important to call `Validate` on `CachetMonitor` and all the monitors inside.

[API Documentation](https://godoc.org/github.com/conradoqg/cachet-monitor/cachet)

# Contributions welcome

We'll happily accept contributions for the following (non exhaustive list).

- Implement ICMP check
- Implement TCP check
- Any bug fixes / code improvements
- Test cases

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details