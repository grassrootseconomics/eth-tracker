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

See [docker-compose.yaml](dev/docker-compose.yaml) for an example on how to run and deploy a single instance.

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

For an example, see `dev/docker-compose.nats.yaml`.

### 3. Update config values

See `.env.example` on how to override default values defined in `config.toml` using env variables. Alternatively, mount your own config.toml either during build time or Docker runtime.

```bash
# Override only specific config values
nano .env.example
mv .env.example .env
```

Special env variables:

* DEV=*
* FORCE_BACKFILL=*

Refer to [`config.toml`](config.toml) to understand different config value settings.


### 4. Run the tracker

```bash
cd dev
docker compose up
```

## Processing NATS messages

### JSON structure

```js
{
    "block": Number,
    "contractAddress": String,
    "success": Boolean,
    "timetamp" Number,
    "transactionHash": String,
    "transactionType": String,
    "payload": Object
}
```

### Monitoring with NATS CLI

Install NATS CLI from [here](https://github.com/nats-io/natscli?tab=readme-ov-file#installation).

```bash
nats subscribe "TRACKER.*"
```

## Caveats

* Reverted transactions older than 10 minutes will be skipped due to the trie potentially missing. To override this behavior, use an archive node and set `chain.archive_node` to `true`.
* The backfiller will only re-queue an epoch's (17,280 blocks) worth of missing blocks. To override this behavior, pass an environment variable `FORCE_BACKFILL=*`. This may lead to degraded RPC performance.

## License

[AGPL-3.0](LICENSE).
