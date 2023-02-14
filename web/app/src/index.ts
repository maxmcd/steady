import { client } from "twirpscript";
import { Login } from "./steady.pb";

import * as monaco from "monaco-editor";
console.log("iasdfi");
self.MonacoEnvironment = {
  getWorkerUrl: function (moduleId, label) {
    if (label === "json") {
      return "./assets/language/json/json.worker.js";
    }
    if (label === "css" || label === "scss" || label === "less") {
      return "./assets/language/css/css.worker.js";
    }
    if (label === "html" || label === "handlebars" || label === "razor") {
      return "./assets/language/html/html.worker.js";
    }
    if (label === "typescript" || label === "javascript") {
      return "./assets/language/typescript/ts.worker.js";
    }
    return "./assets/editor/editor.worker.js";
  },
};

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

  let runApplicationForm = document.forms["run-application" as any];
  if (runApplicationForm) {
    runApplicationForm.onsubmit = function () {
      let element = runApplicationForm.elements[
        "index.ts" as any
      ] as HTMLInputElement;
      element.value = editor.getValue();
    };
  }
})();
