
# Running the exporter

    ./veeam-ahvproxy-exporter -proxy.url https://1.2.3.4:8100 -proxy.username <user> -proxy.password <password>

    curl localhost:9405/metrics

# Running exporter with different sections

    veeam-ahvproxy-exporter -exporter.conf ./config.yml

During the Query pass GET-Parameter Section

    curl localhost:9405/metrics?section=cluster01


Config
```
default:
  url: https://1.2.3.4:8100
  uusername: prometheus
  password: p@ssw0rd

proxy02:
  url: url: https://5.6.7.8:8100
  username: prometheus
  password: qwertz
```

# Prometheus extendended Configuration

Nutanix Config:
```
proxy01:
  url: https://1.2.3.4:8100
  uusername: prometheus
  password: p@ssw0rd

proxy02:
  url: url: https://5.6.7.8:8100
  username: prometheus
  password: qwertz
```

Prometheus Config:
```
scrape_configs:
  - job_name: veeam_ahvproxy
    metrics_path: /metrics
    static_configs:
    - targets:
      - proxy01
      - proxy02
    relabel_configs:
    - source_labels: [__address__]
      target_label: __param_section
    - source_labels: [__address__]
      target_label: __param_target
    - source_labels: [__param_target]
      target_label: instance
    - target_label: __address__
      replacement: ahvproxy_exporter:9405
```

