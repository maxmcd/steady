import { client } from "twirpscript";
import { GetUser } from "../steady.pb";

client.baseURL = "http://localhost:8080";

(async () => {
  const hat = await GetUser({});

  console.log(hat);
})();
