# celo-tracker

![GitHub Tag](https://img.shields.io/github/v/tag/grassrootseconomics/celo-tracker)

A fast and lightweight tracker designed to monitor the Celo blockchain for live and historical transaction events, including reverted transactions. It filters these events and publishes them to NATS for further processing.

It applies deduplication at the NATS level, making it safe to run in a distributed fashion, thus allowing for high availability.

## Getting Started

## Caveats

* Reverted transactions older than 10 minutes will be skipped due to the trie potentially missing. To override this behavior, use an archive node and set `chain.archive_node` to `true`.
* The backfiller will only re-queue an epoch's (17,280 blocks) worth of missing blocks. To override this behavior, pass an environment variable `FORCE_BACKFILL=*`. This may lead to degraded RPC performance.

## License

[AGPL-3.0](LICENSE).
