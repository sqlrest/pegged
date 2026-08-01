# pegged

CLI for managing local PostgreSQL databases in Docker — a **thin shell** over [gomatic/go-pgdocker](https://github.com/gomatic/go-pgdocker): ALL lifecycle behavior (port policy, volume selection, readiness, snapshots, init runs) lives in go-pgdocker; pegged owns only flags/env (`PEGGED_*`, plus `PGPORT`), argument resolution, and JSON output. Never re-implement lifecycle logic here — extend the library instead.

- **Layout** follows the gomatic CLI template standard: strict three-tier `internal/app/commands/<cmd>` (flags → Config binding) → `internal/domain/<cmd>` (`Run(ctx, logger, cfg, args...)` orchestration) → go-pgdocker. Shared plumbing (Manager assembly, port/flag-value resolution) is `internal/domain/manage`; sentinels are `errs.Const` values in `internal/constants`.
- **Testing**: every domain test drives a `pgdocker.Manager` over a package-local fake `Engine` (no daemon, no shared test-helper packages); command tests stub the `runAction` seam. Seam-reassigning tests stay serial and restore via `t.Cleanup`.
- Gate: `make check` (stickler/yze strict, gocognit ≤ 7, exactly 100.0% coverage), run with `GOWORK=off GOPRIVATE=github.com/gomatic`. Managed files (Makefile, .golangci.yaml, workflows) are distributed by `nicerobot/tools.repository` — never hand-edit them.
- **Zero prior-employer trace**: before every push, confirm the working tree and git history carry no prior-employer identifiers — the identifier set lives in the private home policy, not in this public repo. Nothing may match.
