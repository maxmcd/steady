import { client } from "twirpscript";
import { Login } from "../steady.pb";

import * as monaco from "monaco-editor";

client.baseURL = "http://localhost:8080";

(async () => {
  const container = document.getElementById("container");
  if (!container) {
    throw new Error("container element not found");
  }
  const editor = monaco.editor.create(container, {
    value: ["function x() {", '\tconsole.log("Hello world!");', "}"].join("\n"),
    language: "typescript",
  });

  // editor.getModel()?.onDidChangeContent((event) => {
  //   console.log(event);
  // });
  const hat = await Login({ email: "hi", username: "hi" });

  console.log(hat);
})();
