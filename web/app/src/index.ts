import { client } from "twirpscript";
import { Login } from "./steady.pb";

import * as monaco from "monaco-editor";

self.MonacoEnvironment = {
  getWorkerUrl: function (moduleId, label) {
    if (label === "json") {
      return "/assets/language/json/json.worker.js";
    }
    if (label === "css" || label === "scss" || label === "less") {
      return "/assets/language/css/css.worker.js";
    }
    if (label === "html" || label === "handlebars" || label === "razor") {
      return "/assets/language/html/html.worker.js";
    }
    if (label === "typescript" || label === "javascript") {
      return "/assets/language/typescript/ts.worker.js";
    }
    return "/assets/editor/editor.worker.js";
  },
};

client.baseURL = "http://localhost:8080";

(async () => {
  const editorContainer = document.getElementById("monaco-editor");
  if (!editorContainer) {
    throw new Error("container element not found");
  }
  let text = editorContainer.innerText;
  editorContainer.innerText = "";
  const editor = monaco.editor.create(editorContainer, {
    value: text,
    language: "typescript",
    minimap: {
      enabled: false,
    },
  });
  monaco.languages.typescript.javascriptDefaults.setCompilerOptions({
    lib: ["ESNext"],
    moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
    strict: true,
    allowJs: true,
  });
  let resp = await fetch("/assets/bun-types/types.d.ts");
  monaco.languages.typescript.typescriptDefaults.addExtraLib(
    await resp.text(),
    "bun-types"
  );
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
