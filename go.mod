module veeam-ahvproxy-exporter

replace (
  github.com/claranet/veeam_ahvproxy => ./internal/veeam
)

require (
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/yaml.v2 v2.4.0
	github.com/claranet/veeam_ahvproxy v0.0.0
)
