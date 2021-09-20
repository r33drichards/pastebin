<!DOCTYPE html>
<html>
  <head>
    <title>PBIN pastebin with Monaco Editor</title>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link
      href="https://unpkg.com/tailwindcss@^2/dist/tailwind.min.css"
      rel="stylesheet"
    />
    <link
      rel="stylesheet"
      data-name="vs/editor/editor.main"
      href="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.20.0/min/vs/editor/editor.main.min.css"
    />
  </head>
  <body>
    <div class="container-xl h-screen overflow-y-hidden">
      <div class="max-h-1/5 p-4">
        <a class="py-2 px-4 font-bold text-black" href="/"> PBIN </a>
        <button
          class="
            py-2
            px-4
            font-semibold
            rounded-lg
            shadow-md
            text-white
            bg-green-500
            hover:bg-grey-700
          "
          onclick="navigator.clipboard.writeText(window.location)"
        >
          share url
        </button>
        <button
          class="
            py-2
            px-4
            font-semibold
            rounded-lg
            shadow-md
            text-white
            bg-green-500
            hover:bg-green-700
          "
          onclick="window.copyText()"
        >
          copy text
        </button>
      </div>
      <div id="monacoContainer" class="w-full h-4/5"></div>
    </div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.26.1/min/vs/loader.min.js"></script>
    <script>
      // require is provided by loader.min.js.
      require.config({
        paths: {
          vs: "https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.26.1/min/vs",
        },
      });
      require(["vs/editor/editor.main"], () => {
        var m = monaco.editor.create(
                document.getElementById("monacoContainer"),
                {
                  value: {{ .Text }},
                  language: {{ .Language }},
                  theme: "vs-dark",
                  readOnly: true
                }
        );
        window.copyText = function () {
          navigator.clipboard.writeText(
                  m.getValue()
          );
        }
      });
    </script>
  </body>
</html>
