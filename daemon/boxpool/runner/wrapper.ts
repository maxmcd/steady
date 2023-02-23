

let toLoad = process.env.STEADY_INDEX_LOCATION
// TODO: delete when bun support delete
process.env.STEADY_INDEX_LOCATION = undefined
let healthEndpoint = process.env.STEADY_HEALTH_ENDPOINT
// TODO: delete when bun support delete https://github.com/oven-sh/bun/issues/1559
process.env.STEADY_HEALTH_ENDPOINT = undefined

let module = require(toLoad)

let app = module.default;
if (app === undefined) {
    throw new Error("Invalid module export, exiting")
}
if (app.fetch == undefined) {
    throw new Error("Invalid module export, exiting")
}

const AsyncFunction = (async () => {}).constructor;

app.port = process.env.PORT
app.development = false

let innerFetch = app.fetch;
if (innerFetch instanceof AsyncFunction) {
    app.fetch = (async (request: Request): Promise<Response> => {
        if (request.url.endsWith(healthEndpoint)) {
            return new Response("steady")
        }
        return innerFetch(request)
    })
} else {
    app.fetch = (request: Request): Response => {
        if (request.url.endsWith(healthEndpoint)) {
            return new Response("steady")
        }
        return innerFetch(request)
    }
}
export default app
