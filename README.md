# uWSGI expoter

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
## Configure uWSGI expoter

Use as example config.yaml
Fist 2 paramether is enougth clear.

soket_dir: folder where is stored all stats soket.
Than expoter will join soket_dir with soket
In this example:
All uWSGI stats soket are in /run/uwsgi/stats/, so uWSGI expoter will read, [your_soket_name.sock, other_soket_domain.sock]
``` yaml
port:9237
pidfile: "/run/uwsgi_expoter.pid"

soket_dir: "/run/uwsgi/stats/"
stats_sokets:

- domain: your.domain.com
  soket: your_soket_name.sock

- domain: other-domain.com
  soket: other_soket_domain.sock
```

