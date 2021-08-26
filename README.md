# uWSGI exporter

## Installation

You can install uWSGI exporter in system folder.
Usually I use `/opt/prometheus/...`
Then create relative SystemD unit.

Create prometheus user, or run in your preferred user, nobody?.
``` bash
adduser prometheus --system --no-create-home --shell /sbin/nologin
```

Systemd example file:
``` systemd
[Unit]
Description=uWSGI exporter
After=syslog.target

[Service]
Type=simple
PermissionsStartOnly=true
GuessMainPID=true
WorkingDirectory=/opt/prometheus/uwsgi_exporter

PIDFile=/run/uwsgi_exporter.pid
ExecStartPre=/bin/touch /run/uwsgi_exporter.pid
ExecStartPre=/bin/chown prometheus:prometheus /run/uwsgi_exporter.pid

ExecStart=/opt/prometheus_suite/uwsgi_exporter/uWSGI_exporter

[Install]
WantedBy=multi-user.target
```
## Configure uWSGI exporter

Use as example config.yaml
First 2 paramether is clear enough.

`socket_dir`: folder where all stats sockets are stored.
Then exporter will join `socket_dir` with `socket`
In this example:
All uWSGI stats sockets are in `/run/uwsgi/stats/`, so uWSGI exporter will read, `[your_socket_name.sock, other_socket_domain.sock]`

If you use full path both will use it without join.
``` yaml
port:9237
pidfile: "/run/uwsgi_exporter.pid"

socket_dir: "/run/uwsgi/stats/"
stats_sockets:

- domain: your.domain.com
  socket: your_socket_name.sock

- domain: other-domain.com
  socket: other_socket_domain.sock
```

