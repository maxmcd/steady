# 2023-05-01

Minimum features:

1. Web editor to edit bun
2. Run the thing you're editing, get a link to it in browser
3. Add additional files to the project.
4. Import other projects?


# 2023-03-08

Type safe bun server lib! https://elysiajs.com/

Consider just patching bun to be safe?

- [Filesystem stuff for node-scoped libs](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/node/types.zig#L547-L548)
- [Would also need to patch direct uses of the filesystem, eg sqlite](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/bindings/sqlite/sqlite.exports.js#L175)
- [OS file should be audited](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/node/node_os.zig#L14-L15)
- [No syscalls](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/node/syscall.zig)
- [Module imports can access the OS](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/module_loader.zig)
- Most explicit node modules that are candidates for removal [could be removed here](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/module_loader.zig#L1655)
- [JavaScriptCore](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/bun-jsc.exports.js#L1-L2) seems safe, is it?
- [Need to be sure node:path is stubbed](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/path.exports.js#L7)
- [Gotta make sure wasi keeps working](https://github.com/oven-sh/bun/blob/deb7a2b19225265ad8f847da300f9f6db7c5e8b3/src/bun.js/wasi.exports.js#L771-L773)
- Sqlite seems to be the only non-node import that uses the filesystem. Make sure that's the case?
- How can we test this? Possible to run a subset of Bun tests and then confirm that syscalls match what we expect?

Could also just remove os/syscall/net/dns/http/child_process/wasi entirely and deal with adding them when needed. Just use sqlite for now.

# 2023-03-01

Resources for getting some version of git hosting working for the feature below:


Example of creating a git server:
- https://github.com/go-git/go-git/issues/234
- https://github.com/seankhliao/gitreposerver


A brief foray into pluggable fs layers when thinking about how to store this stuff.
- https://pkg.go.dev/github.com/go-git/go-billy/v5
- https://github.com/hack-pad/hackpadfs
- https://github.com/spf13/afero



# 2023-02-28

Applications could import like so:
```ts
import Foo from "steady.sh/maxmcd/foo"
```
Which would add the following to the package.json:
```json
{"steady.sh/maxmcd/foo": "git://steady.sh/maxmcd/foo.git#v1.0.27"}
```

We can then allow imports from other projects and get TS definitions linked between projects.

Our helper methods could also do things like this:

```ts
import counterServer from "steady.sh/maxmcd/counter"
let app = Steady.runApplication("steady.sh/maxmcd/counter");
const client = Steady.client(counterServer, app);

console.log(await client.Increment())
```

A contrived example of something that might look like a type safe server/client
setup. The Magic here is that Steady.runApplication will use the version from
package.json to determine which application version to run.


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


