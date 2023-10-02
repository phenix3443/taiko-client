# Differences from the official repository

- [x] Add vscode setting. If you use vscode as an IDE, you will benefit in many ways：
  - [debugging](https://code.visualstudio.com/docs/editor/debugging), you can use breakpoint to make debugging more easy.
  - [tasks](https://code.visualstudio.com/docs/editor/tasks)
  - [extensions], some recommend extension from my workspace, may be some of them are surprises.
  - One-click startup of the debugging environment eliminates the need to manually prepare the test network.
  - and more!
- [x] Move go bindings to [forked taiko-mono repo](https://github.com/phenix3443/taiko-mono)
- [x] The unit test execution logic has been refactored with docker sdk, allowing unit tests to be executed in parallel, twice as fast as before.
- [x] Optimize command line flag parsing based on flag.Destination and [flag.Action](https://cli.urfave.org/v2/examples/flags/#flag-actions), removing a lot of invalid copy code.
- [ ] Remove over-engineered retry logic based on backoff package.
