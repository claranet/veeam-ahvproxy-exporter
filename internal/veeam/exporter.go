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
	"fmt"

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
	log.Tracef("Convert '%v'(%T) to float64", value, value)

	var v float64
	switch value.(type) {
	case int:
	  v = float64(value.(int))
		break
	case float64:
		v = value.(float64)
		break
	default:
		v, _ = strconv.ParseFloat(value.(string), 64)
		break
	}

	return v
}

func (e *ahvProxyExporter) dateToUnixTimestamp(value string) float64 {

	layout := "1/2/2006 3:04:05 PM"
	t, _ := time.Parse(layout, value)
	log.Tracef("Convert '%s' to Timestamp '%d'",value, t.Unix())
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
				Help:      "Number of Protected VMs on the Cluster"},
				[]string{})
	e.metrics["protected_vms"].Describe(ch)

	e.metrics["unprotected_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "unprotected_vms",
				Help:      "Number of Unprotected vms on the Cluster"},
				[]string{})
	e.metrics["unprotected_vms"].Describe(ch)

	e.metrics["total_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "total_vms",
				Help:      "Number of total vms on the Cluster"},
				[]string{})
	e.metrics["total_vms"].Describe(ch)

	e.metrics["snapshot_protected_vms"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "snapshot_protected_vms",
				Help:      "Number of VMs protected by snapshot"},
				[]string{})
	e.metrics["snapshot_protected_vms"].Describe(ch)

	e.metrics["job_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_count",
				Help:      "Numnber of Jobs Managed by the Proxy"},
				[]string{})
	e.metrics["job_count"].Describe(ch)

	jobs_labels := []string{"job_id", "job_name"}
	e.metrics["job_state"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_state",
				Help:      "Status of the Backup Job - 0 - Success, 1 - Warning, 3 - Error, 4 - Unknown"},
				jobs_labels)
	e.metrics["job_state"].Describe(ch)

	e.metrics["job_vms_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_vms_count",
				Help:      "Number of VMs Protected by this Job"},
				jobs_labels)
	e.metrics["job_vms_count"].Describe(ch)

	e.metrics["job_next_run"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_next_run",
				Help:      "Unix Timestamp of next scheduled run"},
				jobs_labels)
	e.metrics["job_next_run"].Describe(ch)

	e.metrics["job_last_run"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_last_run",
				Help:      "Unix Timestamp of last run"},
				jobs_labels)
	e.metrics["job_last_run"].Describe(ch)

	e.metrics["job_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_status",
				Help:      "Status of the job"},
				append(jobs_labels, "job_status"))
	e.metrics["job_status"].Describe(ch)

	e.DescribeJobVm(ch)
}

func (e *ahvProxyExporter) DescribeJobVm(ch chan<- *prometheus.Desc) {
	job_vm_labels := []string{"job_id", "job_name", "vm_id", "vm_name"}
	e.metrics["job_vm_last_success"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_vm_last_success",
				Help:      ""},
				job_vm_labels)
	e.metrics["job_vm_last_success"].Describe(ch)

	e.metrics["job_vm_restore_points"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_vm_restore_points",
				Help:      ""},
				job_vm_labels)
	e.metrics["job_vm_restore_points"].Describe(ch)

	e.metrics["job_vm_size_bytes"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: e.namespace,
				Name:      "job_vm_size_bytes",
				Help:      ""},
				job_vm_labels)
	e.metrics["job_vm_size_bytes"].Describe(ch)
}

// Collect - Implemente prometheus.Collector interface
// See https://github.com/prometheus/client_golang/blob/master/prometheus/collector.go
func (e *ahvProxyExporter) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start Collecting ...")

	var protstatus map[string]interface{}
	resp, _ := e.api.makeRequest("GET", "/api/v4/dashboard/protectedVms")
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&protstatus)
	log.Debugf("Prection state: %v", protstatus)

	g := e.metrics["protected_vms"].WithLabelValues()
	g.Set(protstatus["protectedVms"].(float64))
	g.Collect(ch)
	g  = e.metrics["unprotected_vms"].WithLabelValues()
	g.Set(protstatus["unprotectedVms"].(float64))
	g.Collect(ch)
	g  = e.metrics["snapshot_protected_vms"].WithLabelValues()
	g.Set(protstatus["protectedVmsWithSnapshots"].(float64))
	g.Collect(ch)
	g  = e.metrics["total_vms"].WithLabelValues()
	g.Set(protstatus["totalVms"].(float64))
	g.Collect(ch)

	var result map[string]interface{}
	resp, _ = e.api.makeRequest("GET", "/api/v4/jobs/")
	data := json.NewDecoder(resp.Body)
	data.Decode(&result)

	g = e.metrics["job_count"].WithLabelValues()
	g.Set( e.valueToFloat64( result["totalCount"] ) )
	g.Collect(ch)

	vmStack := make(map[string]map[string]string)
	var resultList []interface{} = result["results"].([]interface{})
	for _, j := range resultList {
		job := j.(map[string]interface{})

		var jobId, jobName string = job["id"].(string), job["name"].(string)

		g := e.metrics["job_vms_count"].WithLabelValues(jobId, jobName)
		g.Set( e.valueToFloat64( job["objects"] ) )
		g.Collect(ch)

		g = e.metrics["job_last_run"].WithLabelValues(jobId, jobName)
		g.Set(e.dateToUnixTimestamp( job["lastRunUtc"].(string) ))
		g.Collect(ch)

		// Iterate over vm ids in job settings and
		// push them to the stack
		// this results in a mapping taple of vmid => ( job )
		var settings map[string]interface{} = job["settings"].(map[string]interface{})
		for _, id := range settings["vmIds"].([]interface{}) {
			vmStack[id.(string)] = map[string]string{"jobName": jobName, "jobId": jobId}
		}

		var startTimestamp, jobState float64
		startTime := job["nextRunUtc"].(string)
		startTimestamp = 0
		jobState = 0
		if startTime != "Disabled" {
			startTimestamp = e.dateToUnixTimestamp(startTime)
			jobState = 1
		}
		g = e.metrics["job_next_run"].WithLabelValues(jobId, jobName)
		g.Set(startTimestamp)
		g.Collect(ch)
		
		g = e.metrics["job_state"].WithLabelValues(jobId, jobName)
		g.Set(jobState)
		g.Collect(ch)

		status := strings.ToLower( job["status"].(string) )
		value := 0
		g = e.metrics["job_status"].WithLabelValues(jobId, jobName, status)
		switch status {
		  case "success": value = 0
			default: value = 1
		}
		g.Set(e.valueToFloat64(value))
		g.Collect(ch)

		// // e.CollectJobVms(policy["id"].(string), policy["name"].(string), ch)
	}

	var clusters []map[string]interface{}
	resp, _ = e.api.makeRequest("GET", "/api/v4/clusters/")
	decoder = json.NewDecoder(resp.Body)
	decoder.Decode(&clusters)

	for _, cluster := range clusters {
		var clusterId string = cluster["id"].(string)

		var result map[string]interface{}
		resp, _ = e.api.makeRequest("GET", fmt.Sprintf("/api/v4/clusters/%s/vms", clusterId))
		decoder = json.NewDecoder(resp.Body)
		decoder.Decode(&result)

		var resultList []interface{} = result["results"].([]interface{})
		for _, v := range resultList {
			vm := v.(map[string]interface{})
			var jobId = vmStack[vm["id"].(string)]["jobId"]
			var jobName = vmStack[vm["id"].(string)]["jobName"]
			if jobId == "" {
				continue
			}
			g = e.metrics["job_vm_size_bytes"].WithLabelValues(jobId, jobName, vm["id"].(string), vm["name"].(string))
			g.Set(vm["vmSize"].(float64))
			g.Collect(ch)
		}

		resp, _ = e.api.makeRequest("GET", "/api/v4/protectedVms")
		decoder = json.NewDecoder(resp.Body)
		decoder.Decode(&result)
		resultList = result["results"].([]interface{})
		for _, v := range resultList {
			vm := v.(map[string]interface{})
			var jobId = vmStack[vm["id"].(string)]["jobId"]
			var jobName = vmStack[vm["id"].(string)]["jobName"]
			if jobId == "" {
				continue
			}
			g := e.metrics["job_vm_restore_points"].WithLabelValues(jobId, jobName, vm["id"].(string), vm["name"].(string))
			g.Set(vm["backups"].(float64))
			g.Collect(ch)

			g = e.metrics["job_vm_last_success"].WithLabelValues(jobId, jobName, vm["id"].(string), vm["name"].(string))
			g.Set(e.dateToUnixTimestamp(vm["lastProtectionDateUtc"].(string)))
			g.Collect(ch)
		}
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
