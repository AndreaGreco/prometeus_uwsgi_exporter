# uWSGI exporter

## Installation

You can install uWSGI expoter in system folder.
Usually i use, /opt/prometheus/...
Then create relative SystemD unit.

Create prometheus user, or run in your preferred user, nobody?.
``` bash
adduser prometheus --system --no-create-home --shell /sbin/nologin
```

Systemd example file:
``` systemd
[Unit]
Description=uWSGI expoter
After=syslog.target

[Service]
Type=simple
PermissionsStartOnly=true
GuessMainPID=true
WorkingDirectory=/opt/prometheus/uwsgi_expoter

PIDFile=/run/uwsgi_expoter.pid
ExecStartPre=/bin/touch /run/uwsgi_expoter.pid
ExecStartPre=/bin/chown prometheus:prometheus /run/uwsgi_expoter.pid

ExecStart=/opt/prometheus_suite/uwsgi_expoter/uWSGI_expoter

[Install]
WantedBy=multi-user.target
```
## Configure uWSGI exporter

Use as example config.yaml
Fist 2 paramether is enougth clear.

socket_dir: folder where is stored all stats socket.
Than expoter will join socket_dir with socket
In this example:
All uWSGI stats socket are in /run/uwsgi/stats/, so uWSGI expoter will read, [your_socket_name.sock, other_socket_domain.sock]

If you use full path both will use it without join.
``` yaml
port:9237
pidfile: "/run/uwsgi_expoter.pid"

socket_dir: "/run/uwsgi/stats/"
stats_sockets:

- domain: your.domain.com
  socket: your_socket_name.sock

- domain: other-domain.com
  socket: other_socket_domain.sock
```

