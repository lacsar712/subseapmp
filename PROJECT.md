# subseapmp — project notes

Subsea pump manifold booster station for multiphase production tie-back.

## Architecture

```
cmd/subseapmp          CLI entry
internal/app           orchestrates station FSM, schedules, snapshots
internal/manifold      headers, slots, routing
internal/pump          booster plant, flow/pressure control
internal/pipeline      sensors, hold windows, segment path planner
internal/fsm           station + booster state machines
internal/interlock     valve locks, routing guards
internal/store         memory store, schedules, snapshots
internal/clock         process + wall clocks
internal/config        station configuration
internal/model         domain types
internal/alarms        alarm registry and emitter
```

## Operational flow

1. Prime manifold headers and validate flow setpoints
2. Start booster units via coordinator FSM
3. Arm pipeline pressure hold window
4. Release hold after process clock elapses
5. Persist station snapshot to store
