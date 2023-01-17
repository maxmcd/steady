# Steady

- HTTP request/response
- Worker memory and cpu limited (KVM?)
- Backup sqlite and restore when needed
- Web framework that exposes typescript annotations when available
- Exposes http routing (websockets?) at the very least
- Package.json and libraries
- Webassembly?


## Temporal

- https://legacy-documentation-sdks.temporal.io/go/how-to-create-a-worker-session-in-go
- [MaxConcurrentWorkflowTaskExecutionSize](https://github.com/temporalio/sdk-go/blob/a080d2c5313465efdc350892bf20ed814ad8addd/internal/worker.go#L84)
- [DeadlockDetectionTimeout](https://github.com/temporalio/sdk-go/blob/a080d2c5313465efdc350892bf20ed814ad8addd/internal/worker.go#L168)
- https://docs.temporal.io/tasks#task-routing
- https://github.com/temporalio/samples-go/tree/main/fileprocessing

### Arch thoughts

- Daemon on the machine that handles all requests and managing state
- Temporal worker handles registering jobs and rebalancing.
- Temporal worker on each host handles bookkeeping.

Lifecycle:
- Upload new application
- Gets added to hash-ring and sent to correct worker workflow
- Workflow downloads needed files and set up host daemon to run invocations
- Load balance gets update on location

Slicer workload:
- Sees new hosts come online, gives them a slice range.
- Can be queried for the current host assignments/keys, assigns a queue to each host for workflow/task assignment

Worker workload:
- Starts the daemon?
- Sees each new job and downloads needed files
- Handles migrations?
