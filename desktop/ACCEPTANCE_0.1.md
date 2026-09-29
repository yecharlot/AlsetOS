# Alset Desktop 0.1 — Acceptance

| # | Test | Status |
|---|------|--------|
| 1 | Boots in VM from USB/ISO | pending |
| 2 | Loads Tiny Core + required extensions | pending |
| 3 | Starts X / compatible display server | pending |
| 4 | Opens Alset Shell automatically | partial (dev: `go run` bridge + shell) |
| 5 | Shows desktop, menu, task area | partial (shell UI scaffold) |
| 6 | Opens a real application | pending |
| 7 | Starts AlsetOS as independent service | partial (bridge spawns/calls `alsetos`) |
| 8 | Shows organism status | partial (API `/v1/organism`) |
| 9 | Keeps config/data across reboot | pending (TC `tce=` notes) |
| 10 | Recovery mode if shell fails | partial (boot scripts document FLWM fallback) |

**0.1 goal:** reproducible path, not a finished ISO in every environment.
