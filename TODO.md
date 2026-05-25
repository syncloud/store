# TODO

## Move the top-level store module into its own `store/` subdirectory

The repo root currently holds the main Go module (`module github.com/syncloud/store`) alongside the `test/` and `verify/` sub-modules. So the root has `cmd/`, `api/`, `model/`, `storage/`, `release/`, `rest/`, `crypto/`, `internal/`, `log/`, `machine/`, `util/`, `web/`, `build.sh`, `Dockerfile`, etc. — plus the sibling `ci/`, `config/`, `deploy/`, `test/`, `verify/` directories.

Move the store binary's own tree under `store/`:

```
/  (root, just orchestration)
├── ci/
├── config/
├── deploy/
├── test/
├── verify/
├── store/                  ← go.mod here, all store source moves under
│   ├── api/
│   ├── cmd/
│   ├── model/
│   ├── storage/
│   ├── release/
│   ├── rest/
│   ├── crypto/
│   ├── internal/
│   ├── log/
│   ├── machine/
│   ├── util/
│   ├── web/
│   ├── build.sh
│   ├── Dockerfile
│   └── Dockerfile.store-publisher
├── .drone.jsonnet
└── README.md
```

Touch points to update:
- `build.sh` paths (output dir, dockerfile context)
- `.drone.jsonnet` (build / docker step working dir, `dockerfile:` settings)
- `Dockerfile` and `Dockerfile.store-publisher` `COPY build/bin/...` lines
- `test/test.sh` paths to scp deploy/ etc.
- README / docs
