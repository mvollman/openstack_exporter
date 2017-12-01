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

package collectors

import (
  "database/sql"
  // Overriding sql pakage
  _ "github.com/go-sql-driver/mysql"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/common/log"
)

// A InstanceUsageCollector is used to gather all the current resource 
// allocations to instances by tenant across an Openstack cloud
type InstanceUsageCollector struct {
  conn *sql.DB 

  // Instances shows the allocated instances
  Instances *prometheus.GaugeVec

  // Vcpus shows the allocated vcpus
  Vcpus *prometheus.GaugeVec

  // Memory shows the allocated memory
  Memory *prometheus.GaugeVec

  // LocalDisk shows the allocated local disk
  LocalDisk *prometheus.GaugeVec

}

// NewInstanceUsageCollector creates and returns the reference to InstanceUsageCollector
// and internally defines each metric.
func NewInstanceUsageCollector(conn *sql.DB, cloud string) *InstanceUsageCollector {
  labels := make(prometheus.Labels)
  labels["cloud"] = cloud
  instanceLabels := []string{"project"}

  return &InstanceUsageCollector{
    conn: conn,

    Instances: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "instances_total", 
      Help:        "Openstack Instances",
      ConstLabels: labels,
    },
    instanceLabels,
    ),
    Vcpus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "vcpus_total", 
      Help:        "Openstack VCPUs",
      ConstLabels: labels,
    },
    instanceLabels,
    ),
    Memory: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "memory_mb_total", 
      Help:        "Openstack Memory",
      ConstLabels: labels,
    },
    instanceLabels,
    ),
    LocalDisk: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "root_gb_total", 
      Help:        "Openstack Local Disk",
      ConstLabels: labels,
    },
    instanceLabels,
    ),
  }
}

func (c *InstanceUsageCollector) metricList() []*prometheus.GaugeVec {
  return []*prometheus.GaugeVec{
    c.Instances,
    c.Vcpus,
    c.Memory,
    c.LocalDisk,
  }
}

func (c *InstanceUsageCollector) collect() error {
  rows, err := c.conn.Query(`
    SELECT count(c.vcpus) as instances,
      sum(c.vcpus) as vcpus,
      sum(c.memory_mb) as memory_mb,
      sum(c.root_gb) as root_gb,
      d.name as project
    FROM nova.instances c
    JOIN keystone.project d ON c.project_id = d.id 
     AND c.deleted_at IS NULL
    GROUP BY d.name;
  `)

  if err != nil {
    log.Fatal(err.Error())
  }

  for rows.Next() {
    var instances, vcpus, memoryMb, rootGb float64
    var project string
    err = rows.Scan(&instances,
                    &vcpus,
		    &memoryMb,
		    &rootGb,
		    &project)
    if err != nil {
      log.Fatal(err.Error())
    }
    c.Instances.WithLabelValues(project).Set(instances)
    c.Vcpus.WithLabelValues(project).Set(vcpus)
    c.Memory.WithLabelValues(project).Set(memoryMb)
    c.LocalDisk.WithLabelValues(project).Set(rootGb)
  }
  if err = rows.Err(); err != nil {
    log.Fatal(err.Error())
  }
  return nil

}

// Describe sends the descriptors of each metric over the provided channel.
// The correspinding metric values are sent separately.
func (c *InstanceUsageCollector) Describe(ch chan<- *prometheus.Desc) {
  for _, metric := range c.metricList() {
    metric.Describe(ch)
  }
}

// Collect sends the metric values for each metric pertaining to the global
// instance usage over to the provided prometheus Metric channel.
func (c *InstanceUsageCollector) Collect(ch chan<- prometheus.Metric) {
  if err := c.collect(); err != nil {
    log.Errorln("Failed collecting openstack instance metrics:", err)
    return
  }

  for _, metric := range c.metricList() {
    metric.Collect(ch)
  }
}
