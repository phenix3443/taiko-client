> [Differences](./differ.md) from the official repository. If you like clean, beautiful code, you'll love this fork.
>
# taiko-client

[![Run Test](https://github.com/phenix3443/taiko-client/actions/workflows/test.yml/badge.svg)](https://github.com/phenix3443/taiko-client/actions/workflows/test.yml)
[![Docker Image](https://github.com/phenix3443/taiko-client/actions/workflows/docker.yml/badge.svg)](https://github.com/phenix3443/taiko-client/actions/workflows/docker.yml)

Taiko protocol's client software implementation in Go. Learn more about Taiko nodes with [the docs](https://taiko.xyz/docs/concepts/taiko-nodes).

## Project structure

| Path                | Description                                                                                                                              |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `bindings/`         | [Go contract bindings](https://geth.ethereum.org/docs/dapp/native-bindings) for Taiko smart contracts, and few related utility functions |
| `cmd/`              | Main executable for this project                                                                                                         |
| `docs/`             | Documentation                                                                                                                            |
| `driver/`           | Driver sub-command                                                                                                                       |
| `integration_test/` | Scripts to do the integration testing of all client softwares                                                                            |
| `metrics/`          | Metrics related                                                                                                                          |
| `pkg/`              | Library code which used by all sub-commands                                                                                              |
| `proposer/`         | Proposer sub-command                                                                                                                     |
| `prover/`           | Prover sub-command                                                                                                                       |
| `scripts/`          | Helpful scripts                                                                                                                          |
| `testutils/`        | Test utils                                                                                                                               |
| `version/`          | Version information                                                                                                                      |

## Build the source

Building the `taiko-client` binary requires a Go compiler. Once installed, run:

```sh
make build
```

## Usage

Review all available sub-commands:

```sh
bin/taiko-client --help
```

Review each sub-command's command line flags:

```sh
bin/taiko-client <sub-command> --help
```

## Testing

Ensure you have Docker running, and pnpm installed.

Then, run the integration tests:

1. Start Docker locally
2. Perform a `pnpm install` in `taiko-mono/packages/protocol`
3. Replace `<PATH_TO_TAIKO_MONO_REPO>` and execute:

   ```bash
   TAIKO_MONO_DIR=<PATH_TO_TAIKO_MONO_REPO> \
   COMPILE_PROTOCOL=true \
     make test
   ```
