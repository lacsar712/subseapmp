# subseapmp

Subsea pump manifold booster station coordinating multiphase boosters, production header routing, pipeline pressure holds, and slot-level manifold allocation.

## Domain

- **Manifold**: production headers, slot tables, jumper routing between tree nodes
- **Pump**: booster coordination, flow validation, suction/discharge pressure control
- **Pipeline**: pressure sensors, hold windows, riser segment registry and path planning
- **FSM**: station lifecycle (idle → priming → boosting → hold) and per-booster states
- **Interlock**: valve locks and slot-to-manifold guards
- **Store**: in-memory snapshots and boost schedules
- **Clock**: process clock for deterministic subsea operation windows

## Build

```bash
make build
make test
```

## Run

```bash
go run ./cmd/subseapmp
```

## Benzhi

See `BENZHI_README.md` and `build_benzhi_docker.sh`.
