# open-source-path

A command-line tool that tracks trending repositories in the cloud-native ecosystem using the GitHub Search API. Results are saved as daily markdown reports and committed to this repository automatically.

## What it does

- Queries GitHub for trending repositories across cloud-native topics (Kubernetes, CNCF, service mesh, Go)
- Deduplicates results across multiple queries
- Generates a structured markdown report with stars, language, and description
- Archives one report per day under `reports/` and keeps `reports/latest.md` up to date

## Installation

Requires Go 1.21 or later.

```bash
git clone https://github.com/siddubakka/open-source-path.git
cd open-source-path
go build -o oss-radar .
```

## Usage

```bash
# Run with GitHub token (recommended — raises rate limit from 60 to 5000 req/hour)
GITHUB_TOKEN=your_token ./oss-radar

# Run without token (unauthenticated, 60 req/hour limit)
./oss-radar
```

The tool writes output to:
- `reports/YYYY-MM-DD.md` — dated archive
- `reports/latest.md` — always the most recent report

## Automation

A GitHub Actions workflow runs the tool daily and commits any new reports. See `.github/workflows/daily.yml`.

## Project structure

```
.
├── main.go
├── go.mod
├── internal/
│   ├── fetcher/
│   │   └── github.go       queries the GitHub Search API
│   └── report/
│       └── generator.go    builds and writes the markdown report
├── reports/
│   ├── latest.md
│   └── YYYY-MM-DD.md
└── .github/
    └── workflows/
        └── daily.yml
```

## Configuration

The queries tracked are defined in `internal/fetcher/github.go`. Edit the `queries` slice to change what topics are tracked.

## License

MIT
