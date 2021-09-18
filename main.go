package main

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"net/http"
)

var (
	PBIN_TABLE_NAME = os.Getenv("PBIN_TABLE_NAME")
	PBIN_URL        = os.Getenv("PBIN_URL")
)

// Item defines the item for the table
// snippet-start:[dynamodb.go.get_item.struct]
type Paste struct {
	PK       string
	SK       string
	Language string
	Text     string
}

//https://stackoverflow.com/questions/2377881/how-to-get-a-md5-hash-from-a-string-in-golang
func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GetTables retrieves a list of your Amazon DynamoDB tables
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     limit is the maximum number of tables to return
// Output:
//     If success, a list of the tables and nil
//     Otherwise, nil and an error from the call to ListTables
func GetTables(sess *session.Session, limit *int64) ([]*string, error) {
	// snippet-start:[dynamodb.go.list_all_tables.call]
	svc := dynamodb.New(sess)

	result, err := svc.ListTables(&dynamodb.ListTablesInput{
		Limit: limit,
	})
	// snippet-end:[dynamodb.go.list_all_tables.call]
	if err != nil {
		return nil, err
	}

	return result.TableNames, nil
}

// MakeTable creates an Amazon DynamoDB table
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     attributeDefinitions describe the table's attributes
//     keySchema defines the table schema
//     tableName is the name of the table
// Output:
//     If success, nil
//     Otherwise, an error from the call to CreateTable
func MakeTable(svc dynamodbiface.DynamoDBAPI, attributeDefinitions []*dynamodb.AttributeDefinition, keySchema []*dynamodb.KeySchemaElement, tableName *string) error {

	_, err := svc.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: attributeDefinitions,
		KeySchema:            keySchema,
		TableName:            tableName,
		BillingMode:          aws.String(dynamodb.BillingModePayPerRequest),
	})
	return err
}

// GetTableItem retrieves the item with the year and title from the table
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     table is the name of the table
//     title is the movie title
//     year is when the movie was released
// Output:
//     If success, the information about the table item and nil
//     Otherwise, nil and an error from the call to GetItem or UnmarshalMap
func GetTableItemPK(svc dynamodbiface.DynamoDBAPI, table, pk *string) (*Paste, error) {

	keyCond := expression.Key("PK").Equal(expression.Value(pk))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)
	expr, err := builder.Build()
	if err != nil {
		panic(err)
	}
	queryInput := dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 aws.String(PBIN_TABLE_NAME),
	}

	result, err := svc.Query(&queryInput)
	// snippet-end:[dynamodb.go.get_item.call]
	if err != nil {
		return nil, err
	}

	// snippet-start:[dynamodb.go.get_item.unmarshall]
	if result.Items == nil {
		msg := "Could not find '" + *pk + "'"
		return nil, errors.New(msg)
	}

	item := Paste{}

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &item)
	// snippet-end:[dynamodb.go.get_item.unmarshall]
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// AddTableItem adds an item to an Amazon DynamoDB table
// Inputs:
//     sess is the current session, which provides configuration for the SDK's service clients
//     year is the year when the movie was released
//     table is the name of the table
//     title is the movie title
//     plot is a summary of the plot of the movie
//     rating is the movie rating, from 0.0 to 10.0
// Output:
//     If success, nil
//     Otherwise, an error from the call to PutItem
func AddTableItem(svc dynamodbiface.DynamoDBAPI, table, text, hash, lang *string, trys int) (*string, error) {
	if trys == 0 {
		return nil, errors.New("trys exceeded")
	}
	// snippet-start:[dynamodb.go.create_new_item.assign_struct]
	currentTime := time.Now()

	item := Paste{
		PK:       *hash,
		SK:       currentTime.Format("2006-01-06"),
		Text:     *text,
		Language: *lang,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	// snippet-end:[dynamodb.go.create_new_item.assign_struct]
	if err != nil {
		return nil, err
	}

	// snippet-start:[dynamodb.go.create_new_item.call]
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:                av,
		TableName:           table,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// process SDK error
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return AddTableItem(svc, table, text, aws.String(GetMD5Hash(*hash)), lang, trys-1)
			} else {
				// todo err 500
				panic(err)
			}
		}
	}

	return hash, nil
}

func init() {
	sess := session.Must(session.NewSession())

	limit := int64(10)
	tables, err := GetTables(sess, aws.Int64(limit))
	if err != nil {
		fmt.Println("Got an error retrieving table names:")
		fmt.Println(err)
		return
	}

	// create dynamodb table if it doesn't exist
	createTable := true
	for _, n := range tables {
		if 0 == strings.Compare(*n, PBIN_TABLE_NAME) {
			createTable = false
			break
		}
	}
	svc := dynamodb.New(sess)

	if createTable {

		attributeDefinitions := []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: aws.String("S"),
			},
		}

		keySchema := []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       aws.String("RANGE"),
			},
		}

		go MakeTable(svc, attributeDefinitions, keySchema, aws.String(PBIN_TABLE_NAME))
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
		sess := session.Must(session.NewSession())
		svc := dynamodb.New(sess)
		switch request.Method {
		case "POST":
			if err := request.ParseForm(); err != nil {
				fmt.Fprintf(writer, "ParseForm() err: %v", err)
				return
			}
			text := request.FormValue("text")
			lang := request.FormValue("lang")
			hash := GetMD5Hash(text)
			id, err := AddTableItem(svc, aws.String(PBIN_TABLE_NAME), aws.String(text), aws.String(hash), aws.String(lang), 10)

			if err != nil {
				// todo 500 err
				panic(err)
			}
			q := request.URL.Query()
			q.Del("text")
			q.Del("lang")
			q.Set("id", *id)
			request.URL.RawQuery = q.Encode()
			http.Redirect(writer, request, request.URL.String(), 301)
		case "GET":
			id := request.URL.Query().Get("id")
			log.Println(id)
			paste, err := GetTableItemPK(svc, aws.String(PBIN_TABLE_NAME), aws.String(id))
			if err != nil {
				// TODO return 500 err
				panic(err)
			}
			lang := paste.Language
			text := paste.Text
			// warning! this is unsafe
			fmt.Fprintf(writer, `<!DOCTYPE html>
<html>
<head>

	<link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/themes/prism.min.css" rel="stylesheet" />
</head>
<body>
<button onclick="copyText()">Copy!</button>
<pre><code id="paste" class="language-%s">%s</code></pre>

	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/prism.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.25.0/components/prism-%s.min.js" ></script>
</body>
<script>
function copyText(){
    navigator.clipboard.writeText(document.getElementById("paste").textContent)
}
</script>
</html>`, lang, text, lang)

		default:
			http.Redirect(writer, request, PBIN_URL, 301)

		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(`<!doctype html>
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
	<option value="objective->objective-c</option>c"
	<option value="pascal">pascal</option>
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
</html>`))
		if err != nil {
			log.Println(err)
		}
	})
	log.Println("server listening")
	http.ListenAndServe(":8000", nil)
}
