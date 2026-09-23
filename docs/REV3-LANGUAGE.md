# Craft 0.1.3 — Timer and task contracts

Implementation is available in source. Build and user acceptance are recorded
separately in [Rev.3 status](../roadmap/phase1/output/phase1-rev3-status.md).
Rev.2 remains the baseline; this document specifies additions and migration.

## API

| API | Return | Contract |
| --- | --- | --- |
| `std.task.spawn(worker, args...)` | `Task` | Schedule a named user function returning `Void` |
| `std.timer.after(delay, worker, args...)` | `Timer` | Invoke once after a nonnegative `Duration` |
| `std.timer.every(interval, worker, args...)` | `Timer` | Repeat with a positive `Duration` |
| `std.time.sleep(duration)` | `Void` | Suspend the calling task; negative duration is an error |
| `std.time.sleep(milliseconds)` | `Void` | Existing Int overload, 0 through 86400000 inclusive |
| `handle.join()` | `Void` | Wait for completion; rethrow failure or `CancellationError` |
| `handle.cancel()` | `Void` | Request cooperative cancellation; idempotent |
| `handle.status()` | `String` | Snapshot: pending/running/completed/failed/cancelled |

The callback reference must be a bare function identifier, not `worker()` or a
variable containing a function. Arguments are positional and checked against
the callback signature. General function values, closures, async/await keywords,
callback return values and message channels are outside this release.
`Task` and `Timer` are built-in types; existing structs with these names must be renamed.

```craft
func worker(label: String) {
    std.time.sleep(std.time.durationMilliseconds(10))
    print(label)
}

func main() {
    let a: Task = std.task.spawn(worker, "one")
    let b: Task = std.task.spawn(worker, "two")
    a.join()
    b.join()
    let t: Timer = std.timer.after(std.time.durationMilliseconds(20), worker, "timer")
    t.join()
}
```

## Values, ownership and errors

Arguments are evaluated once, left to right, in the caller and copied at
submission. Each invocation receives new deep copies, including each repeating
timer callback. No lexical captures or shared mutable Craft variables are added.
Task/Timer handles are exceptions to value-copy semantics: copies, parameters,
arrays, maps and struct fields still reference the same underlying operation.
They cannot be converted to JsonValue; print displays their kind and status.
Cancellation is permitted through a `let` handle, just as joining is.

The owning scope is the **function invocation** (or test invocation) that
created the handle, not an inner brace block and not the variable's lifetime.
On scope exit, all owned operations are cancelled and joined **before** that
function's deferred blocks run. Returning a handle does not transfer ownership:
the returned handle already refers to a finished/cancelled operation.
Call `join()` before leaving the function if the operation must complete.
There are no detached tasks, background jobs after main, or cross-test tasks.

Joining a failed operation marks its failure observed and propagates its
exception. Repeated joins give the same terminal outcome. Unobserved failures
propagate to the owning function at exit in creation order; an existing body
error takes precedence. Scope-driven cancellation alone does not fail the owner.
Joining a cancelled task throws `CancellationError`, code `E3004`.
Cancellation of the *joining task itself* unwinds it instead of being catchable
as another task's cancellation. Self/ancestor joins and explicit wait cycles
raise catchable runtime errors.

Cancellation checks occur at expression/statement/loop boundaries and during
sleep, timer waits, joins and console waits. Deferred cleanup is attempted after
cancellation with a five-second window per unwinding function; nested cleanup
calls inherit that window. Normal, uncancelled defer has no new time limit.
Starting tasks/timers from defer, including indirectly through helpers, raises
a runtime error. OS calls that cannot be interrupted can delay cleanup beyond
the cooperative deadline. Existing body errors take precedence over cleanup errors.

## Scheduling and timer policy

The runtime uses bounded Go goroutines with independent Craft call stacks and
environments, synchronized handles/output, and periodic scheduler yields.
It supports concurrency without promising CPU parallel speedup or task ordering.
Limits are 256 active operations per run and 256 owned handles per function
invocation (including completed handles until that invocation ends).
Spawning beyond either limit raises a catchable error. Ordinary sequential
helper invocations can create and finish their own operations independently.

Waiting uses Go's monotonic timers. Delays are minimum waits, not real-time
deadlines. `sleep(0)` yields; `after(0, ...)` dispatches asynchronously and may
run before the caller's next statement. Timer creation never invokes the
callback inline. Negative waits and zero/negative repeating intervals fail.

`every` uses **fixed delay**: first wait, callback, wait again. A slow callback
cannot overlap another callback of the same timer or accumulate missed ticks.
Separate timers can run concurrently. Callback failure stops the timer.
Pending cancellation prevents dispatch; running cancellation is cooperative.
Cancel/dispatch are synchronized. Completion and cancellation can race: a
completed operation remains completed; a running operation observes cancellation
at its next check. Cancel is a request, so use join to await cleanup.

## Input, output and limits

Each print/write call is serialized; ordering between tasks is unspecified.
Only one console read may be outstanding; overlapping reads raise a runtime
error. Cancelling a read ends the Craft wait and makes that session's input
unavailable for subsequent reads. A raw blocking reader may still own one worker
until it returns: embedders must close/unblock their custom reader, and process
exit reclaims CLI resources. It is not a detached Craft task.
File operations remain synchronous, and same-file access must be coordinated
by the program. Copying arguments does not duplicate external resources.

No cron, calendar schedule, persistent schedule, restart recovery, OS wakeup,
mutex/channel API or general event library is provided.

## Compatibility and acceptance

Existing Int sleep and Rev.2 syntax remain supported. Parser and formatter use
the existing named-call syntax; the type checker resolves callback references
only at the three task/timer call sites. Bundles remain format 2 and now carry
language `0.1.3`. This CLI reads 0.1.1/0.1.2/0.1.3 bundles; older CLIs must not be
used for new bundles. Extension version 0.1.0 is independent of CLI version.

User-run checks: [TRY-REV3](TRY-REV3.md). Runtime tests include a manually
advanced clock to check concurrent waits, fixed delay, one-shot invocation and
cancellation without timing assertions about real elapsed milliseconds.
