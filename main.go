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
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"net/http"
)

var (
	//go:embed templates/index.html.tpl
	INDEX_TEMPLATE_TEXT string
	//go:embed templates/paste.html.tpl
	PASTE_TEMPLATE_TEXT string
	PBIN_TABLE_NAME     = os.Getenv("PBIN_TABLE_NAME")
	PBIN_URL            = os.Getenv("PBIN_URL")
)

// Item defines the item for the table
// snippet-start:[dynamodb.go.get_item.struct]
type Paste struct {
	PK       string
	SK       string
	Language string
	Text     string
}

type PasteTemplateContent struct {
	Text, Language string
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
			ptc := PasteTemplateContent{
				text,
				lang,
			}
			t := template.Must(template.New("paste").Parse(PASTE_TEMPLATE_TEXT))
			err = t.ExecuteTemplate(writer, "paste", ptc)
			if err != nil {
				// TODO return 500
				panic(err)
			}

		default:
			http.Redirect(writer, request, PBIN_URL, 301)

		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(INDEX_TEMPLATE_TEXT))
		if err != nil {
			log.Println(err)
		}
	})
	log.Println("server listening")
	http.ListenAndServe(":8000", nil)
}
