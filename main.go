//   Copyright 2017 Mike Vollman
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License

// Command openstack_exporter provides a Prometheus exporter for an Openstack
// cloud.
package main

import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "net/http"
  "sync"

  "./collectors"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/prometheus/common/log"
  "gopkg.in/alecthomas/kingpin.v2"
)

// OpenstackExporter wraps all openstack collectors and provides a single global
// exporter to extract metrics out of.  It also ensures that the collection is
// done in a thread-safe manner, the necessary requirement stated by prometheus.
// It also implements a prometheus.Collector interface in order to register it
// correctly.
type OpenstackExporter struct {
  mu         sync.Mutex
  collectors []prometheus.Collector
}

// Verify that the exporter implements the interface correctly.
var _ prometheus.Collector = &OpenstackExporter{}

// NewOpenstackExporter creates an instance of OpenstackExporter and returns a
// reference to it.  We can choose to enable a collector to extract stats out
// of it by adding it to the list of collectors.
func NewOpenstackExporter(conn *sql.DB, cloud string) *OpenstackExporter {
  return &OpenstackExporter{
    collectors: []prometheus.Collector{
      collectors.NewQuotaUsageCollector(conn, cloud),
      collectors.NewInstanceUsageCollector(conn, cloud),
    },
  }
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (o *OpenstackExporter) Describe(ch chan<- *prometheus.Desc) {
  for _, oc := range o.collectors {
    oc.Describe(ch)
  }
}
// Collect sends all the descriptors of the collectors included to the provided
// channel.
func (o *OpenstackExporter) Collect(ch chan<- prometheus.Metric) {
  o.mu.Lock()
  defer o.mu.Unlock()

  for _, oc := range o.collectors {
    oc.Collect(ch)
  }
}

func main() {
  var (
    listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9100").String()
    metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
    dbAddress     = kingpin.Flag("db.address", "Address of Openstack DB server to Query").String()
    dbPort        = kingpin.Flag("db.port", "Port of Openstack DB server to Query").Default("3306").String()
    dbUsername    = kingpin.Flag("db.username", "Username of Openstack DB server to Query").String()
    dbPassword    = kingpin.Flag("db.password", "Password of Openstack DB server to Query").String()
    cloud         = kingpin.Flag("os.cloud", "Cloud label to assign to metrics").Default("default").String()
  )
  log.AddFlags(kingpin.CommandLine)
  kingpin.HelpFlag.Short('h')
  kingpin.Parse()

  myDsn := *dbUsername + ":" + *dbPassword + "@tcp(" + *dbAddress + ":" + *dbPort + ")/"
  conn, err := sql.Open("mysql", myDsn)
  if err != nil {
    log.Fatal(err.Error())
  }
  defer conn.Close()
  prometheus.MustRegister(NewOpenstackExporter(conn, *cloud))

  log.Infoln("Starting openstack_exporter")
  http.Handle(*metricsPath, promhttp.Handler())
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`<html>
            <head><title>Openstack Exporter</title></head>
	    <body>
	    <h1>Openstack Exporter</h1>
	    <p><a href='` + *metricsPath + `'>Metrics</a></p>
	    </body>
	    </html>`))
  })

  log.Infoln("Listening on", *listenAddress)
  log.Fatal(http.ListenAndServe(*listenAddress, nil))

}
