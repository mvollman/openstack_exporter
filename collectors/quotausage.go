package collectors

import (
  "database/sql"
  // Overriding sql pakage
  _ "github.com/go-sql-driver/mysql"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/common/log"
)

const (
  openstackNamespace = "openstack"
)

// A QuotaUsageCollector is used to gather all the current quota settings per
// tenant across an Openstack cloud
type QuotaUsageCollector struct {
  conn *sql.DB 

  // VcpuQuota shows the VCPU Quota
  VcpuQuota *prometheus.GaugeVec

  // GigabytesQuota shows the Gigabytes Quota
  GigabytesQuota *prometheus.GaugeVec

  // InjectedFileContentQuota shows the InjectedFileContent Quota
  InjectedFileContentQuota *prometheus.GaugeVec

  // InjectedFilesQuota shows the InjectedFiles Quota
  InjectedFilesQuota *prometheus.GaugeVec

  // InstancesQuota shows the Instances Quota
  InstancesQuota *prometheus.GaugeVec

  // MetadataQuota shows the Metadata Quota
  MetadataQuota *prometheus.GaugeVec

  // RAMQuota shows the RAM Quota
  RAMQuota *prometheus.GaugeVec

  // SnapshotsQuota shows the Snapshots Quota
  SnapshotsQuota *prometheus.GaugeVec

  // VolumesQuota shows the Volumes Quota
  VolumesQuota *prometheus.GaugeVec

}

// NewQuotaUsageCollector creates and returns the reference to QuotaUsageCollector
// and internally defins each metric that display quotas.
func NewQuotaUsageCollector(conn *sql.DB, cloud string) *QuotaUsageCollector {
  labels := make(prometheus.Labels)
  labels["cloud"] = cloud
  quotaLabels := []string{"project"}

  return &QuotaUsageCollector{
    conn: conn,

    VcpuQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "vcpus_quota", 
      Help:        "Openstack VCPUs Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    GigabytesQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "vcpus_quota", 
      Help:        "Openstack VCPUs Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    InjectedFileContentQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "inject_file_content_bytes_quota", 
      Help:        "Openstack Injected file content bytes Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    InjectedFilesQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "injected_files_quota", 
      Help:        "Openstack Injected Files Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    InstancesQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "instances_quota", 
      Help:        "Openstack Instances Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    MetadataQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "metadata_bytes_quota", 
      Help:        "Openstack Metadata bytes Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    RAMQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "ram_quota", 
      Help:        "Openstack RAM Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    SnapshotsQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "snapshots_quota", 
      Help:        "Openstack Snapshots Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
    VolumesQuota: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace:   openstackNamespace,
      Name:        "volumes_quota", 
      Help:        "Openstack Volumes Quota",
      ConstLabels: labels,
    },
    quotaLabels,
    ),
  }
}

func (c *QuotaUsageCollector) metricList() []*prometheus.GaugeVec {
  return []*prometheus.GaugeVec{
    c.VcpuQuota,
    c.GigabytesQuota,
    c.InjectedFileContentQuota,
    c.InjectedFilesQuota,
    c.InstancesQuota,
    c.MetadataQuota,
    c.RAMQuota,
    c.SnapshotsQuota,
    c.VolumesQuota,
  }
}

func (c *QuotaUsageCollector) collect() error {
  rows, err := c.conn.Query(`
    SELECT a.resource,
      case when a.hard_limit = -1 then 0 else a.hard_limit end as hard_limit,
      b.name as project
    FROM cinder.quotas a
    JOIN keystone.project b ON a.project_id = b.id AND a.deleted_at IS NULL 
    UNION ALL 
    SELECT a.resource, 
      case when a.hard_limit = -1 then 0 else a.hard_limit end as hard_limit,
      b.name as project
    FROM nova.quotas a 
    JOIN keystone.project b ON a.project_id = b.id AND a.deleted_at IS NULL;
  `)

  if err != nil {
    log.Fatal(err.Error())
  }

  for rows.Next() {
    var quota string
    var qval float64
    var project string
    err = rows.Scan(&quota, &qval, &project)
    if err != nil {
      log.Fatal(err.Error())
    }
    switch quota {
    case "cores":
      c.VcpuQuota.WithLabelValues(project).Set(qval)
    case "gigabytes":
      c.GigabytesQuota.WithLabelValues(project).Set(qval)
    case "injected_file_content_bytes":
      c.InjectedFileContentQuota.WithLabelValues(project).Set(qval)
    case "injected_files":
      c.InjectedFilesQuota.WithLabelValues(project).Set(qval)
    case "instances":
      c.InstancesQuota.WithLabelValues(project).Set(qval)
    case "metadata_items":
      c.MetadataQuota.WithLabelValues(project).Set(qval)
    case "ram":
      c.RAMQuota.WithLabelValues(project).Set(qval)
    case "snaphots":
      c.SnapshotsQuota.WithLabelValues(project).Set(qval)
    case "volumes":
      c.VolumesQuota.WithLabelValues(project).Set(qval)
    }
  }
  if err = rows.Err(); err != nil {
    log.Fatal(err.Error())
  }
  return nil

}

// Describe sends the descriptors of each metric over the provided channel.
// The correspinding metric values are sent separately.
func (c *QuotaUsageCollector) Describe(ch chan<- *prometheus.Desc) {
  for _, metric := range c.metricList() {
    metric.Describe(ch)
  }
}

// Collect sends the metric values for each metric pertaining to the global
// quota usage over to the provided prometheus Metric channel.
func (c *QuotaUsageCollector) Collect(ch chan<- prometheus.Metric) {
  if err := c.collect(); err != nil {
    log.Errorln("Failed collecting openstack quota metrics:", err)
    return
  }

  for _, metric := range c.metricList() {
    metric.Collect(ch)
  }
}
