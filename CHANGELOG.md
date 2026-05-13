# Changelog

## 0.1.0 (2026-05-13)


### Features

* add list, show, deprecate, and supersede subcommands ([622e3b7](https://github.com/wbern/adr-lint/commit/622e3b7c04cfbdc271a3e71b0c434777ae394a8f))
* add validate, accept/reject/withdraw, and on-disk template ([077ef4f](https://github.com/wbern/adr-lint/commit/077ef4fbbc906c9562866df8cee7d69a64a691f0))
* add version subcommand and arity checks ([2147bfe](https://github.com/wbern/adr-lint/commit/2147bfe953bdfb467419d902834b049a1e788b1e))
* **cli:** add help and unknown-subcommand handling ([f8c70d4](https://github.com/wbern/adr-lint/commit/f8c70d41049c0a1b3df3421dd9098ec5c78d02b6))
* **cli:** add per-subcommand --help and guard self-supersession ([7d55857](https://github.com/wbern/adr-lint/commit/7d558577f4ee30f56f302271cdb768605ceaf00e))
* **cli:** print cwd-relative paths in status messages ([e1ac321](https://github.com/wbern/adr-lint/commit/e1ac3213371529d37110ba2214327d545176f6a7))
* **create:** add create subcommand and seed first ADR ([ba683bc](https://github.com/wbern/adr-lint/commit/ba683bc44bc1e7ddc02376f7088cb9ca54e3e6a9))
* parse superseded_by and surface it in list output ([c9404b7](https://github.com/wbern/adr-lint/commit/c9404b7bb8eccc6d032565cba5b9270183e4c94d))
* **validate:** collect all issues and detect malformed frontmatter ([138664d](https://github.com/wbern/adr-lint/commit/138664d352024516c04d9186d01907d6dbd56247))
* **version:** support build-time version injection via ldflags ([5a22a86](https://github.com/wbern/adr-lint/commit/5a22a86f54513fd37297cfba8674533b00a31d10))


### Bug Fixes

* make ADR writes atomic and create race-free ([47bbd6e](https://github.com/wbern/adr-lint/commit/47bbd6eeb2e824d00af42ea9b5060b183a842057))


### Refactoring

* harden subcommand dispatch and ADR status helpers ([7fb14aa](https://github.com/wbern/adr-lint/commit/7fb14aaee107610a1b49d39fdc8677119e9feeb7))


### Documentation

* add CONTRIBUTING.md with setup and commit conventions ([484b2a7](https://github.com/wbern/adr-lint/commit/484b2a7953c0febe9783498a698c509515bf2b86))
* add recorded demos and restructure README ([eb08ef1](https://github.com/wbern/adr-lint/commit/eb08ef186f68fa6bfe4559dec3aa36876c5a6919))
