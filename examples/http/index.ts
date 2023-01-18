const port = process.env.PORT ?? 3000;
console.log(`Listening on port ${port}`);

export default {
  port,
  fetch(request: Request): Response {
    console.log(`${request.url}`);
    return new Response("Hello Steady");
  },
};
