# Dockerman
What a weird name! I know.
Tird of Kubernetes? Try this. This is a simple Docker comtainer orchestration tool.

## Features
* Multi-backend rotation
* Port-forwarding
* Startup health-check(quick rollback on health check failure)
* No storage required

## Roadmap
* Centralized auto log-collection(in-progress)
* Cross-host communication(based on swarm)

## Prepare Docker Daemoon to Listen on 2375 Port
1. Open up service daemon settings
```shell
sudo systemctl edit docker.service
```

2. Add following lines:
```shell
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H fd:// -H tcp://127.0.0.1:2375
# could be  ExecStart=/usr/sbin/dockerd -H fd:// -H tcp://127.0.0.1:2375 for some occasions
```

3. Reload the systemd manager configuration:
```shell
sudo systemctl daemon-reload
```

4. Restart docker service
```shell
sudo systemctl restart docker
```
