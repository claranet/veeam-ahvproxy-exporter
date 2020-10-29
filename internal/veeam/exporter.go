//
// veeam-ahvproxy-exporter
//
// Prometheus Exporter for VEEAM AHV-Porxy API
//
// Author: Martin Weber <martin.weber@de.clara.net>
// Company: Claranet GmbH
//

package veeam

import (
	"encoding/json"

	"strconv"
	"strings"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

)

type ahvProxyExporter struct {
	api       AHVProxy
	result    map[string]interface{}
	metrics   map[string]*prometheus.GaugeVec
	namespace string
}

// ValueToFloat64 converts given value to Float64
func (e *ahvProxyExporter) valueToFloat64(value interface{}) float64 {
	var v float64
	switch value.(type) {
	case float64:
		v = value.(float64)
		break
	case string:
		v, _ = strconv.ParseFloat(value.(string), 64)
		break
	}

	return v
}

func (e *ahvProxyExporter) dateToUnixTimestamp(value string) float64 {

	layout := "01/02/06 3:04:05 PM"
	t, _ := time.Parse(layout, value)
	log.Debugf("Convert '%s' to Timestamp '%s'",value, t.Unix())
	return float64(t.Unix())
	
}

// NormalizeKey replace invalid chars to underscores
func (e *ahvProxyExporter) normalizeKey(key string) string {
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	key = strings.ToLower(key)

	return key
}

// Describe - Implemente prometheus.Collector interface
// See https://github.com/prometheus/client_golang/blob/master/prometheus/collector.go
func (e *ahvProxyExporter) Describe(ch chan<- *prometheus.Desc) {

	e.metrics["protected_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "protected_vms",
				Help:      ""},
				[]string{})
	e.metrics["protected_vms"].Describe(ch)

	e.metrics["unprotected_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "unprotected_vms",
				Help:      ""},
				[]string{})
	e.metrics["unprotected_vms"].Describe(ch)

	e.metrics["total_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "total_vms",
				Help:      ""},
				[]string{})
	e.metrics["total_vms"].Describe(ch)

	e.metrics["snapshot_protected_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "snapshot_protected_vms",
				Help:      ""},
				[]string{})
	e.metrics["snapshot_protected_vms"].Describe(ch)

	e.metrics["job_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_count",
				Help: ""},
				[]string{})
	e.metrics["job_count"].Describe(ch)

	jobs_labels := []string{"id", "name", "type"}
	e.metrics["job_state"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_state",
				Help: ""},
				jobs_labels)
	e.metrics["job_state"].Describe(ch)

	e.metrics["job_vms_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_vms_count",
				Help: ""},
				jobs_labels)
	e.metrics["job_vms_count"].Describe(ch)

	e.metrics["job_next_run"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_next_run",
				Help: ""},
				jobs_labels)
	e.metrics["job_next_run"].Describe(ch)

	e.metrics["job_last_run"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_last_run",
				Help: ""},
				jobs_labels)
	e.metrics["job_last_run"].Describe(ch)

	e.metrics["job_last_scheduled"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_last_scheduled",
				Help: ""},
				jobs_labels)
	e.metrics["job_last_scheduled"].Describe(ch)

	e.metrics["job_creation_date"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_creation_date",
				Help: ""},
				jobs_labels)
	e.metrics["job_creation_date"].Describe(ch)

	e.metrics["job_modification_date"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_modification_date",
				Help: ""},
				jobs_labels)
	e.metrics["job_modification_date"].Describe(ch)

	e.metrics["job_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_status",
				Help: ""},
				append(jobs_labels, "status"))
	e.metrics["job_status"].Describe(ch)
}

// Collect - Implemente prometheus.Collector interface
// See https://github.com/prometheus/client_golang/blob/master/prometheus/collector.go
func (e *ahvProxyExporter) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start Collecting ...")
	var protstatus map[string]interface{}
	resp, _ := e.api.makeRequest("GET", "/api/v1/Dashboard/protstatus/")
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&protstatus)
	log.Debugf("Prection state: %v", protstatus)

	g := e.metrics["protected_vms"].WithLabelValues()
	g.Set(protstatus["protectedVmsCount"].(float64))
	g.Collect(ch)
	g  = e.metrics["unprotected_vms"].WithLabelValues()
	g.Set(protstatus["notProtectedVmsCount"].(float64))
	g.Collect(ch)
	g  = e.metrics["snapshot_protected_vms"].WithLabelValues()
	g.Set(protstatus["protectedVmsWithSnapshotsCount"].(float64))
	g.Collect(ch)
	g  = e.metrics["total_vms"].WithLabelValues()
	g.Set(protstatus["totalVmsCount"].(float64))
	g.Collect(ch)

	var policies map[string]interface{}
	resp, _ = e.api.makeRequest("GET", "/api/v1/policies/")
	data := json.NewDecoder(resp.Body)
	data.Decode(&policies)

	g = e.metrics["job_count"].WithLabelValues()
	log.Debug(policies["MembersCount"].(float64))
	g.Set(e.valueToFloat64(policies["MembersCount"].(float64)))
	g.Collect(ch)

	for _, p := range policies["Members"].([]interface{}) {
		policy := p.(map[string]interface{})
		urlpath := policy["@odata.id"].(string)
		resp, _ = e.api.makeRequest("GET", urlpath)
		var ent map[string]interface{}

		data = json.NewDecoder(resp.Body)
		data.Decode(&ent)

		// labelValues := []string{ent["Id"].(string), ent["name"].(string)}
		g := e.metrics["job_vms_count"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(e.valueToFloat64(ent["vmsCount"]))
		g.Collect(ch)

		g = e.metrics["job_last_run"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(e.dateToUnixTimestamp(ent["lastRun"].(string)))
		g.Collect(ch)

		var startTimestamp, jobState float64
		startTime := ent["startTime"].(string)
		startTimestamp = 0
		jobState = 0
		if startTime != "Disabled" {
			startTimestamp = e.dateToUnixTimestamp(startTime)
			jobState = 1
		}
		g = e.metrics["job_next_run"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(startTimestamp)
		g.Collect(ch)
		
		g = e.metrics["job_state"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(jobState)
		g.Collect(ch)

		g = e.metrics["job_last_scheduled"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(e.dateToUnixTimestamp(ent["lastStartRun"].(string)))
		g.Collect(ch)

		g = e.metrics["job_creation_date"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(e.dateToUnixTimestamp(ent["creationDate"].(string)))
		g.Collect(ch)

		g = e.metrics["job_modification_date"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string))
		g.Set(e.dateToUnixTimestamp(ent["modificationDate"].(string)))
		g.Collect(ch)

		status := strings.ToLower(ent["status"].(string))
		value := 0
		g = e.metrics["job_status"].WithLabelValues(ent["Id"].(string), ent["name"].(string), ent["type"].(string), status)
		switch status {
		  case "success": value = 0
			default: value = 1
		}
		g.Set(e.valueToFloat64(value))
		g.Collect(ch)
	}
}

// NewHostsCollector
func NewAhvProxyExporter(_api *AHVProxy) *ahvProxyExporter {

	return &ahvProxyExporter{
			api:       *_api,
			metrics:   make(map[string]*prometheus.GaugeVec),
			namespace: "veeam_ahvproxy",
		}
}
