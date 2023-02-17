# 2023-02-15

Next is a focus on application editing and basic observability:

- Edit an existing application
    - Send code to existing daemon
    - Oh, try and run code with copy of the database first, see if it runs? Stretch goal.
    - Do we need to keep multiple versions and do a roll-back on failure?
    - Stop app, attempt to start new one
        - Swap source code out "under-foot", otherwise we have to move databases around?
- Service logs
    - Logs dir, combine stdout/stderr for now
    - New versions append to the log
    - No journaling for now
- CLI/repl
- Actual sandboxing?
- Request log

# 2023-02-14

Things to do:

- Logs and streaming logs to frontend
- Multiple files in editor
- Editing application code and updating (w/ validation)
- Migration of new machine, draining applications
- GC of temporary applications
- Sqlite in daemon
- Design bun daemon
- Crons

# 2023-02-12

Run application

- Submit form with index.ts contents
- Generates random app name
- Apps are anonymous, or they are owned by a user, can be associated with source, or not
- Globally unique names
- Running an application will run the app and take you to a page with metadata
-

# 2023-02-10

- Zig web server
- Bun process spawner
- Build in sandboxed filesystem and disabled network
- Have observability platform
- Thread per process for tcp io_uring load balancer
- litestream go subprocess
- twirp for some coms
- heavy duckdb and sqlite use
- worflow per user backed by sqlite, their session. nickname per session for guests. workflows handle state changes and transactional data, private networking. passwords, encrypted data, workflows take action like selecting a backup to roll back to
- core services are just applications, they write to the same application pages, get the same observability


# 2023-02-03

Loadbalancer gets *.steady.page
Loadbalancer get *.steady.sh

Pool of our load balancers


# 2023-02-02

- request to run an application before deploy
- not durable, will be GC'd, will overwrite on name-conflict

- make request to deploy application
- find host that will house the application
- make request to that host


How will that work?

- Load balancer has public and private port
- Private port will proxy api requests to hosts
- Public will proxy application traffic


# 2023-01-25

Expected application migration timeline:

- Slicer assignments are updated
- List of migrations is produced
- All hosts are notified of pending migrations and migrations begin
- Load balancers are notified of new routing table and they have knowledge of old routing table
- During migration period load balancer sends a HEAD request to confirm application is live at the server before sending, otherwise forwards to new host.
- During period where application is fully down requests will queue at the new host

Consider http2 for load balancer to host communication: https://pkg.go.dev/golang.org/x/net/http2/h2c#example-NewHandler

# 2023-01-24

- Stress-testing seems to show that bun can die and then we are left without a process to kill. Figure out why bun is dying and how to recover.

# 2023-01-23

Rough optimistic sqlite migration path:

- Get request to create new application
- Place application on host
- After first service sleep (or earlier?) check for dbs. If they exist, add them to the dbs we are watching and uploading.
- Get request to move service to new host
- Download s3 files to new host in preparation for restore
- Drain service requests on old host, close litestream process, begin downtime.
- Complete database rebuild on new host, start service.
- Communicate to LB/host/whatever to resume sending requests to the correct location

On-host startup process:

- validate that script doesn't cause an obvious errors
- check for existing databases and download them
- ...


# 2023-01-18

## Walk Thoughts

- Workflow per user
- Workflow for application (handles the migrations!)
- Slicer workload
- Stats are collected per-host by a sqlite javascript job?!
- Versioning is complicated because live service live data

## More

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


# 2023-01-17

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

## Arch thoughts

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


