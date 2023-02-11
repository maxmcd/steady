import { client } from "twirpscript";
import { Login } from "../steady.pb";

client.baseURL = "http://localhost:8080";

(async () => {
  const hat = await Login({ email: "hi", username: "hi" });

  console.log(hat);
})();
