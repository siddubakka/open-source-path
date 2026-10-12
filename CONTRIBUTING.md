# Contributing

## Setup

Requires Go 1.21 or later.

```bash
git clone https://github.com/siddubakka/open-source-path.git
cd open-source-path
go build -o oss-radar .
```

## Checks

Run these before pushing:

```bash
gofmt -l .
go vet ./...
go test ./...
```

## Reports

Reports are generated under `reports/` by the daily workflow.
Do not hand-edit dated reports — change the generator and re-run:

```bash
GITHUB_TOKEN=your_token ./oss-radar -json
```
