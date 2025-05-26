# forbidden-planet



Install Docker Compose manually

```bash
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
mkdir -p $DOCKER_CONFIG/cli-plugins
curl -SL https://github.com/docker/compose/releases/download/v2.36.0/docker-compose-linux-aarch64 -o $DOCKER_CONFIG/cli-plugins/docker-compose

chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
docker compose version
```

Docker model runner on linux
```bash
curl https://get.docker.com/| CHANNEL=test sh -
sudo apt install docker-model-plugin
```

