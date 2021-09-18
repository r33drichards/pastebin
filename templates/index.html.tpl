<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Monaco editor</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">

    <link rel="stylesheet" data-name="vs/editor/editor.main" href="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.20.0/min/vs/editor/editor.main.min.css">
</head>
<body>
<div>
    <label for="language">Choose a language:</label>

    <select name="lang" id="language">
        <option value="">--Please choose an option--</option>
        <option value="clojure">clojure</option>
        <option value="cpp">cpp</option>
        <option value="dockerfile">dockerfile</option>
        <option value="elixir">elixir</option>
        <option value="go">go</option>
        <option value="graphql">graphql</option>
        <option value="hcl">hcl</option>
        <option value="javascript">javascript</option>
        <option value="python">python</option>
    </select>
</div>

<div name="text" id="container" style="height:400px;border:1px solid black;"></div>

<button onclick="window.paste()" >paste!</button>


<script src="https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.26.1/min/vs/loader.min.js"></script>
<script>
    // require is provided by loader.min.js.
    require.config({ paths: { 'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.26.1/min/vs' }});
    require(["vs/editor/editor.main"], () => {
        var m = monaco.editor.create(document.getElementById('container'), {
            value: "",
            language: document.getElementById("language").value,
            theme: 'vs-dark',
        });
        document.getElementById("language").addEventListener("change", function (){
            var model = m.getModel(); // we'll create a model for you if the editor created from string value.
            monaco.editor.setModelLanguage(model, document.getElementById("language").value)
        })

        window.paste = function (){
            data ={
                lang: document.getElementById("language").value,
                text: m.getValue()
            }
            post("/paste", data)
        }
    });
    // https://stackoverflow.com/questions/133925/javascript-post-request-like-a-form-submit
    /**
     * sends a request to the specified url from a form. this will change the window location.
     * @param {string} path the path to send the post request to
     * @param {object} params the parameters to add to the url
     * @param {string} [method=post] the method to use on the form
     */
    function post(path, params, method='post') {

        // The rest of this code assumes you are not using a library.
        // It can be made less verbose if you use one.
        const form = document.createElement('form');
        form.method = method;
        form.action = path;

        for (const key in params) {
            if (params.hasOwnProperty(key)) {
                const hiddenField = document.createElement('input');
                hiddenField.type = 'hidden';
                hiddenField.name = key;
                hiddenField.value = params[key];

                form.appendChild(hiddenField);
            }
        }

        document.body.appendChild(form);
        form.submit();
    }

    // function paste(){
    //     data ={
    //     	lang: document.getElementById("language").value,
    //         text: document.getElementById("container").textContent
    //     }
    //    post("/paste", data)
    // }
</script>
</body>
</html>