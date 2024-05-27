# celo-tracker

![GitHub Tag](https://img.shields.io/github/v/tag/grassrootseconomics/celo-tracker)

A fast and lightweight tracker designed to monitor the Celo blockchain for live and historical transaction events, including reverted transactions. It filters these events and publishes them to NATS for further processing.

It applies deduplication at the NATS level, making it safe to run in a distributed fashion, thus allowing for high availability.

## Getting Started

### Prerequisites

* Git
* Docker
* NATS server
* Access to a Celo RPC node

### 1. Build the Docker image

We provide pre-built images for `linux/amd64`. See the packages tab on Github.

If you are on any other platform:

```bash
git clone https://github.com/grassrootseconomics/celo-tracker.git
cd celo-tracker
docker buildx build --build-arg BUILD=$(git rev-parse --short HEAD) --tag celo-tracker:$(git rev-parse --short HEAD) --tag celo-tracker:latest .
docker images
```

### 2. Run NATS

```bash
cd dev
docker compose up -d
docker ps
```

### 3. Update config values

See `.env.example` on how to override default values defined in `config.toml` using env variables. Alternatively, mount your own config.toml either during build time or Docker runtime.

```bash
# Override only specific config values
nano .env.example
mv .env.example .env
```

Refer to [`config.toml`](config.toml) to understand different config value settings.


### 4. Run the tracker

```bash
docker run --env-file .env -p 127.0.0.1:5001:5001 celo-tracker:latest
```

## Caveats

* Reverted transactions older than 10 minutes will be skipped due to the trie potentially missing. To override this behavior, use an archive node and set `chain.archive_node` to `true`.
* The backfiller will only re-queue an epoch's (17,280 blocks) worth of missing blocks. To override this behavior, pass an environment variable `FORCE_BACKFILL=*`. This may lead to degraded RPC performance.

## License

[AGPL-3.0](LICENSE).
