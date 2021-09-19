<!DOCTYPE html>
<html lang="en">
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
                <label class="font-medium text-gray-700" for="language">Select a language: </label>
                <select
                        class="py-2 px-4 rounded-lg shadow-md"
                        name="lang"
                        id="language"
                >
                    <option value="">text</option>
                    <option value="abap">abap</option>
                    <option value="apex">apex</option>
                    <option value="azcli">azcli</option>
                    <option value="bat">bat</option>
                    <option value="bicep">bicep</option>
                    <option value="cameligo">cameligo</option>
                    <option value="clojure">clojure</option>
                    <option value="coffee">coffee</option>
                    <option value="cpp">cpp</option>
                    <option value="csharp">csharp</option>
                    <option value="csp">csp</option>
                    <option value="css">css</option>
                    <option value="dart">dart</option>
                    <option value="dockerfile">dockerfile</option>
                    <option value="ecl">ecl</option>
                    <option value="elixir">elixir</option>
                    <option value="fsharp">fsharp</option>
                    <option value="go">go</option>
                    <option value="graphql">graphql</option>
                    <option value="handlebars">handlebars</option>
                    <option value="hcl">hcl</option>
                    <option value="html">html</option>
                    <option value="ini">ini</option>
                    <option value="java">java</option>
                    <option value="javascript">javascript</option>
                    <option value="julia">julia</option>
                    <option value="kotlin">kotlin</option>
                    <option value="less">less</option>
                    <option value="lexon">lexon</option>
                    <option value="liquid">liquid</option>
                    <option value="lua">lua</option>
                    <option value="m3">m3</option>
                    <option value="markdown">markdown</option>
                    <option value="mips">mips</option>
                    <option value="msdax">msdax</option>
                    <option value="mysql">mysql</option>
                    <option
                            value="objective->objective-c</option>c"
                    <option
                            value="pascal"
                    >
                        pascal
                    </option>
                    <option value="pascaligo">pascaligo</option>
                    <option value="perl">perl</option>
                    <option value="pgsql">pgsql</option>
                    <option value="php">php</option>
                    <option value="postiats">postiats</option>
                    <option value="powerquery">powerquery</option>
                    <option value="powershell">powershell</option>
                    <option value="pug">pug</option>
                    <option value="python">python</option>
                    <option value="qsharp">qsharp</option>
                    <option value="r">r</option>
                    <option value="razor">razor</option>
                    <option value="redis">redis</option>
                    <option value="redshift">redshift</option>
                    <option value="restructuredtext">restructuredtext</option>
                    <option value="ruby">ruby</option>
                    <option value="rust">rust</option>
                    <option value="sb">sb</option>
                    <option value="scala">scala</option>
                    <option value="scheme">scheme</option>
                    <option value="scss">scss</option>
                    <option value="shell">shell</option>
                    <option value="solidity">solidity</option>
                    <option value="sophia">sophia</option>
                    <option value="sparql">sparql</option>
                    <option value="sql">sql</option>
                    <option value="st">st</option>
                    <option value="swift">swift</option>
                    <option value="systemverilog">systemverilog</option>
                    <option value="tcl">tcl</option>
                    <option value="twig">twig</option>
                    <option value="typescript">typescript</option>
                    <option value="vb">vb</option>
                    <option value="xml">xml</option>
                    <option value="yaml">yaml</option>
                </select>
                <button
                        class="
                        float-right
            py-2
            px-4
            font-semibold
            rounded-lg
            shadow-md
            text-white
            bg-green-500
            hover:bg-green-700
          "
                        onclick="window.paste()"
                >
                    paste!
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
            value: "",
            language: document.getElementById("language").value,
            theme: "vs-dark",
          }
        );
        document
          .getElementById("language")
          .addEventListener("change", function () {
            var model = m.getModel(); // we'll create a model for you if the editor created from string value.
            monaco.editor.setModelLanguage(
              model,
              document.getElementById("language").value
            );
          });
        window.paste = function () {
          data = {
            lang: document.getElementById("language").value,
            text: m.getValue(),
          };
          post("/paste", data);
        };
      });
      // https://stackoverflow.com/questions/133925/javascript-post-request-like-a-form-submit
      /**
       * sends a request to the specified url from a form. this will change the window location.
       * @param {string} path the path to send the post request to
       * @param {object} params the parameters to add to the url
       * @param {string} [method=post] the method to use on the form
       */
      function post(path, params, method = "post") {
        // The rest of this code assumes you are not using a library.
        // It can be made less verbose if you use one.
        const form = document.createElement("form");
        form.method = method;
        form.action = path;

        for (const key in params) {
          if (params.hasOwnProperty(key)) {
            const hiddenField = document.createElement("input");
            hiddenField.type = "hidden";
            hiddenField.name = key;
            hiddenField.value = params[key];

            form.appendChild(hiddenField);
          }
        }

        document.body.appendChild(form);
        form.submit();
      }
    </script>
  </body>
</html>
