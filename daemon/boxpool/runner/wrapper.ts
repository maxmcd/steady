

let toLoad = process.env.STEADY_INDEX_LOCATION
process.env.STEADY_INDEX_LOCATION = undefined

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

if (app.fetch instanceof AsyncFunction) {
    app.fetch = (async (request: Request): Promise<Response> => {
        return app.fetch
    })
} else {
    app.fetch = (request: Request): Response => {
        return app.fetch
    }
}
export default app
