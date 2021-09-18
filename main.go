package main

import (
	_ "embed"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"os"
	"strings"

	"log"
	"net/http"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	PBIN_BUCKET_NAME = os.Getenv("PBIN_BUCKET_NAME")
)

// GetAllBuckets retrieves a list of all buckets.
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
// Output:
//     If success, the list of buckets and nil
//     Otherwise, nil and an error from the call to ListBuckets
func GetAllBuckets(sess *session.Session) (*s3.ListBucketsOutput, error) {
	// snippet-start:[s3.go.list_buckets.imports.call]
	svc := s3.New(sess)

	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	// snippet-end:[s3.go.list_buckets.imports.call]
	if err != nil {
		return nil, err
	}

	return result, nil
}
// MakeBucket creates a bucket.
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     bucket is the name of the bucket
// Output:
//     If success, nil
//     Otherwise, an error from the call to CreateBucket
func MakeBucket(sess *session.Session, bucket *string) error {
	// snippet-start:[s3.go.create_bucket.call]
	svc := s3.New(sess)

	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: bucket,
	})
	// snippet-end:[s3.go.create_bucket.call]
	if err != nil {
		return err
	}

	// snippet-start:[s3.go.create_bucket.wait]
	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: bucket,
	})
	// snippet-end:[s3.go.create_bucket.wait]
	if err != nil {
		return err
	}

	return nil
}



func init(){
				sess := session.Must(session.NewSession())
				buckets, err := GetAllBuckets(sess)
				if err != nil {
					log.Panicln(err)
				}
				createBucket := true
				for _, bucket := range  buckets.Buckets {
					if 0 == strings.Compare(*bucket.Name, PBIN_BUCKET_NAME) {
						createBucket = false
						break
					}
				}

				if createBucket {
					MakeBucket(sess, &PBIN_BUCKET_NAME)
				}

}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc("/paste", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			if err := request.ParseForm(); err != nil {
				fmt.Fprintf(writer, "ParseForm() err: %v", err)
				return
			}
			// warning! this is unsafe
			fmt.Fprintf(writer, `
<!DOCTYPE html>
<html>
<head>

	<link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/themes/prism.min.css" rel="stylesheet" />
</head>
<body>
<pre><code class="language-%s">%s</pre>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/prism.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/components/prism-%s.min.js" ></script>
</body>

</html>
`, request.FormValue("lang"), request.FormValue("text"), request.FormValue("lang"))
		default:
			http.Redirect(writer, request, "https://localhost:8000", 301)

		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(`
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
`))
		if err != nil {
			log.Println(err)
		}
	})
	log.Println("server listening")
	http.ListenAndServe(":8000", nil)
}
