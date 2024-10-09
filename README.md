# Dockerman
What a weird name! I know.
Tird of Kubernetes? Try this. This is a simple Docker comtainer orchestration tool.

## Prepare Docker Daemoon to Listen on 2375 Port
1. Open up service daemon settings
```shell
sudo systemctl edit docker.service
```

2. Add following lines:
```shell
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H fd:// -H tcp://0.0.0.0:2375
```

3. Reload the systemd manager configuration:
```shell
sudo systemctl daemon-reload
```

4. Restart docker service
```shell
sudo systemctl restart docker
```
