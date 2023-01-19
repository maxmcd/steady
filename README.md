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


### Walk Thoughts

- Workflow per user
- Workflow for application (handles the migrations!)
- Slicer workload
- Stats are collected per-host by a sqlite javascript job?!
- Versioning is complicated because live service live data



### More

- Slicer
    - Responsible for keeping track of hosts
    - Handles resizing and migration
    - When resizing:
        - Query for stats to resize
        - determine new mapping, return it
        - send signals to every application that needs to migrate? query state in the case of partial failure?
- Application
    - Place application. Send http request to host to download files and start litestream and register application

- User
- Host daemon
    - Load balancer and distribution