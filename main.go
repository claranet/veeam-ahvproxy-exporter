//
// veeam-ahv-exporter
//
// 
//
// Version: v0.4.0
// Author: Martin Weber <martin.weber@de.clara.net>
// Company: Claranet GmbH
//

package main

import (
	"github.com/claranet/veeam_ahvproxy"

	"flag"
	"net/http"

	"strings"
	"fmt"
	"io/ioutil"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var (
	namespace       = "veeam"
	proxyURL        = flag.String("proxy.url", "", "AHV Proxy URL to connect to API https://1.2.3.4:8100")
	proxyUsername   = flag.String("proxy.username", "veeam", "Proxy Username")
	proxyPassword   = flag.String("proxy.password", "veeam", "Proxy User Password")
	listenAddress   = flag.String("listen-address", ":9760", "The address to lisiten on for HTTP requests.")
	logLevel        = flag.String("log.level", "WARNING", "Set debug level")
	exporterConfig  = flag.String("exporter.conf", "", "Config file for multiple sections")
)

type proxy struct {
	Url      string          `yaml:"url"`
	Username string          `yaml:"username"`
	Password string          `yaml:"password"`
}

func main() {
	flag.Parse()
	level := strings.ToLower(*logLevel)
	switch level {
	case "trace": log.SetLevel(log.TraceLevel)
	case "debug": log.SetLevel(log.DebugLevel)
	case "info": log.SetLevel(log.InfoLevel)
	case "warning": log.SetLevel(log.WarnLevel)
	case "error": log.SetLevel(log.ErrorLevel)
	default: log.SetLevel(log.WarnLevel)
	}
	
	//Use locale configfile
	var config map[string]proxy
	var file []byte
	var err error

	if len(*exporterConfig) > 0 {
		//Read complete Config
		file, err = ioutil.ReadFile(*exporterConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		file = []byte(fmt.Sprintf("default: {url: %s, username: %s, password: %s}", *proxyURL, *proxyUsername, *proxyPassword))
	}

	log.Debugf("Config File:\n%s\n", string(file))
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Config: %v", config)

	//	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		section := params.Get("section")
		if len(section) == 0 {
			section = "default"
		}

		log.Infof("Section: %s", section)
		log.Debug("Create Veeam AHV Exporter instance")

		//Write new Parameters
		if conf, ok := config[section]; ok {
			*proxyURL = conf.Url
			*proxyUsername = conf.Username
			*proxyPassword = conf.Password
		} else {
			log.Errorf("Section '%s' not found in config file", section)
			return
		}
		log.Infof("Host: %s", *proxyURL)

		ahvProxyAPI := veeam.NewVeeamAhvProxy(*proxyURL, *proxyUsername, *proxyPassword)

		registry := prometheus.NewRegistry()
		registry.MustRegister(veeam.NewAhvProxyExporter(ahvProxyAPI))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>VEEAM AHV Proxy Exporter</title></head>
		<body>
		<h1>VEEAM AHV Proxy Exporter</h1>
		<p><a href="/metrics?section=default">Metrics</a></p>
		</body>
		</html>`))
	})

	log.Infof("Starting Server: %s", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
