# Mock libvirtd State Machine

## Overview

The mock libvirtd service simulates realistic domain (VM) lifecycle management through a goroutine-based state machine. Each domain runs its own background state machine that handles transitions, boot delays, and metric updates.

## Architecture

### State Definitions

**User-Requested States** (what clients request):
- `running` - VM is powered on and operating
- `shutoff` - VM is powered off
- `paused` - VM is suspended (CPU paused but memory preserved)

**Intermediate States** (internal transitions):
- `starting` - VM is booting (transient, ~1.5s boot delay)
- `stopping` - VM is shutting down (transient, immediate)

### State Transitions

```
shutoff ──[request: running]──> starting ──[~1.5s delay]──> running
                                                                ↓
                                                            ↓ (pause)
                                                           paused
                                                                ↓
                                                        [request: running]
                                                                ↓
                                                            running
                                                                ↓
                                                    [request: shutoff]
                                                                ↓
                                                            stopping
                                                                ↓
                                                            shutoff
```

Valid transitions:
- `shutoff` → `running` (initiates boot via `starting` intermediate)
- `starting` → `running` (automatic, after ~1.5s)
- `running` → `shutoff` (initiates stop via `stopping` intermediate)
- `running` → `paused` (direct transition)
- `paused` → `running` (direct transition)
- `paused` → `shutoff` (initiates stop via `stopping` intermediate)
- `stopping` → `shutoff` (automatic)

Invalid transitions are rejected with HTTP 409 Conflict.

## Implementation Details

### State Machine Loop

Each domain runs a goroutine with a 100ms ticker that:
1. Checks current vs. desired state
2. Transitions through intermediate states based on constraints
3. Updates metrics (uptime, CPU usage, memory usage)
4. Logs all state transitions

### Metrics

Metrics are calculated in real-time based on elapsed time:

**Uptime**: Seconds since `StartedAt` timestamp
- Only tracks while running or paused
- Resets to 0 when shutoff

**CPU Usage**: Ramps from 10% to 40% over ~5 seconds of runtime
- 0% when shutoff or starting
- Continues accruing when paused

**Memory Usage**: Ramps from 20% to 60% over ~3 seconds of runtime
- 0% when shutoff or starting  
- Continues accruing when paused

### Concurrent Safety

- Per-domain: `sync.RWMutex` protects all mutable state
- Global: Handler-level `sync.RWMutex` protects domain map
- Snapshots: `snapshot()` method provides thread-safe read-only copies

### API Response Behavior

When a state change is requested:
- **Immediate response**: Shows the desired state (what was requested)
- **Actual state**: Transitions asynchronously via state machine
- **Polling**: Clients can GET to see real state after transitions complete

This matches real libvirtd behavior - API acknowledges intent immediately while OS-level operations complete in background.

## Boot Timeline

1. **T=0ms**: Client requests state="running"
   - Response shows state="running" (desired)
   - State machine sets desiredState="running"
   - Actual state still "shutoff"

2. **T≤100ms**: First state machine tick
   - Detects desired≠current (running≠shutoff)
   - Transitions shutoff → starting
   - Logs transition

3. **T≤200ms**: Second state machine tick
   - Still in starting state
   - No action yet

4. **T≤1500ms+**: Boot delay elapsed
   - Transitions starting → running
   - Sets StartedAt timestamp
   - Logs transition
   - Metrics begin accruing

5. **T>1500ms**: Steady state
   - State remains "running"
   - Metrics update each tick
   - VM "alive" until shutdown requested

## Testing

The implementation includes comprehensive tests:

- **TestStateMachineTransitions**: Verifies full boot/pause/stop cycle
- **TestInvalidStateTransition**: Validates rejection of invalid requests
- **TestConcurrentDomains**: 3+ VMs transitioning simultaneously
- **TestMetricsProgression**: Metrics increase over time correctly

Run tests with:
```bash
go test ./internal/handler -v -run "StateMachine|Concurrent|Metrics"
```

## Example Usage

### Start a VM
```bash
curl -X PUT http://localhost:8080/api/domains/{id} \
  -H "Content-Type: application/json" \
  -d '{"state": "running"}'
```

Response (immediate):
```json
{
  "id": "uuid",
  "state": "running",
  "started_at": 1717951234567,
  "uptime": 0,
  "cpu_usage": 0,
  "mem_usage": 0
}
```

After 2 seconds, GET the domain:
```bash
curl http://localhost:8080/api/domains/{id}
```

Response (actual state):
```json
{
  "id": "uuid",
  "state": "running",
  "started_at": 1717951234567,
  "uptime": 2,
  "cpu_usage": 15.5,
  "mem_usage": 35.2
}
```

### Pause a running VM
```bash
curl -X PUT http://localhost:8080/api/domains/{id} \
  -H "Content-Type: application/json" \
  -d '{"state": "paused"}'
```

### Stop a VM
```bash
curl -X PUT http://localhost:8080/api/domains/{id} \
  -H "Content-Type: application/json" \
  -d '{"state": "shutoff"}'
```

## Performance Characteristics

- **Boot delay**: ~1.5 seconds (configurable via `bootTime` constant)
- **State tick rate**: 100ms (configurable via `stateTickRate` constant)
- **Memory per domain**: ~1KB base + goroutine overhead
- **CPU per domain**: Negligible - runs once per tick (100ms interval)
- **Concurrent domains**: Tested with 3+, scales linearly with goroutines
