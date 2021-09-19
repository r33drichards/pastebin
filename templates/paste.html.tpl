<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.css">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/themes/prism.min.css" rel="stylesheet" />
</head>
<body>
<button onclick="copyText()">Copy!</button>
<pre><code id="paste" class="language-{{ .Language }}">{{ .Text }}</code></pre>

<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/prism.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/components/prism-{{ .Language }}.min.js" ></script>
</body>
<script>
    function copyText(){
        navigator.clipboard.writeText(document.getElementById("paste").textContent)
    }
</script>
</html>