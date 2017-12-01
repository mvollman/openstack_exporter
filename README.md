# openstack_exporter
OpenStack DB data exported as Prometheus Metrics

Connect directly to an Openstack DB server using a read only account
to export usage data to prometheus.

```
$ ./openstack_exporter --help
usage: openstack_exporter [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and
                                 --help-man).
      --web.listen-address=":9100"  
                                 Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --db.address=DB.ADDRESS    Address of Openstack DB server to Query
      --db.port="3306"           Port of Openstack DB server to Query
      --db.username=DB.USERNAME  Username of Openstack DB server to Query
      --db.password=DB.PASSWORD  Password of Openstack DB server to Query
      --os.cloud="default"       Cloud label to assign to metrics
      --log.level="info"         Only log messages with the given severity or above. Valid
                                 levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"  
                                 Set the log target and format. Example:
                                 "logger:syslog?appname=bob&local=7" or
                                 "logger:stdout?json=true"
```


