package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

// DataStore is the interface for our database operations
type DataStore interface {
	GetPaste(id string) (*Paste, error)
	AddPaste(text, lang string) (string, error)
	GetDiff(id string) (*Diff, error)
	AddDiff(oldText, newText string) (string, error)
	Close() error
}

// Paste represents a paste item
type Paste struct {
	PK       string
	SK       string
	Language string
	Text     string
}

// Diff represents a diff item
type Diff struct {
	PK      string
	SK      string
	OldText string
	NewText string
}

// BoltStore implements DataStore using BoltDB
type BoltStore struct {
	db *bolt.DB
}

// DynamoStore implements DataStore using DynamoDB
type DynamoStore struct {
	svc       *dynamodb.DynamoDB
	tableName string
}

// NewDataStore creates a new DataStore based on the configuration
func NewDataStore() (DataStore, error) {
	dbType := os.Getenv("DB_TYPE")
	switch dbType {
	case "bolt":
		return NewBoltStore()
	case "dynamo":
		return NewDynamoStore()
	default:
		return NewBoltStore()
	}
}

// NewBoltStore creates a new BoltStore
func NewBoltStore() (*BoltStore, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pbin.db"
	}
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("pastes"))
		if err != nil {
			return fmt.Errorf("create pastes bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("diffs"))
		if err != nil {
			return fmt.Errorf("create diffs bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

// NewDynamoStore creates a new DynamoStore
func NewDynamoStore() (*DynamoStore, error) {
	sess := session.Must(session.NewSession())
	limit := int64(10)
	tables, err := GetTables(sess, aws.Int64(limit))
	if err != nil {
		fmt.Println("Got an error retrieving table names:")
		fmt.Println(err)
		return nil, err
	}

	// create dynamodb table if it doesn't exist
	createTable := true
	for _, n := range tables {
		if strings.Compare(*n, PBIN_TABLE_NAME) == 0 {
			createTable = false
			break
		}
	}
	svc := dynamodb.New(sess)
	tableName := os.Getenv("DYNAMO_TABLE_NAME")
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
	return &DynamoStore{svc: svc, tableName: tableName}, nil
}

// GetTables retrieves a list of your Amazon DynamoDB tables
// Inputs:
//
//	sess is the current session, which provides configuration for the SDK's service clients
//	limit is the maximum number of tables to return
//
// Output:
//
//	If success, a list of the tables and nil
//	Otherwise, nil and an error from the call to ListTables
func GetTables(sess *session.Session, limit *int64) ([]*string, error) {
	svc := dynamodb.New(sess)

	result, err := svc.ListTables(&dynamodb.ListTablesInput{
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	return result.TableNames, nil
}

// MakeTable creates an Amazon DynamoDB table
// Inputs:
//
//	sess is the current session, which provides configuration for the SDK's service clients
//	attributeDefinitions describe the table's attributes
//	keySchema defines the table schema
//	tableName is the name of the table
//
// Output:
//
//	If success, nil
//	Otherwise, an error from the call to CreateTable
func MakeTable(svc dynamodbiface.DynamoDBAPI, attributeDefinitions []*dynamodb.AttributeDefinition, keySchema []*dynamodb.KeySchemaElement, tableName *string) error {
	_, err := svc.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: attributeDefinitions,
		KeySchema:            keySchema,
		TableName:            tableName,
		BillingMode:          aws.String(dynamodb.BillingModePayPerRequest),
	})
	return err
}

// GetPaste retrieves a paste from BoltDB
func (b *BoltStore) GetPaste(id string) (*Paste, error) {
	var paste Paste
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pastes"))
		v := bucket.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("paste not found")
		}
		return json.Unmarshal(v, &paste)
	})
	if err != nil {
		return nil, err
	}
	return &paste, nil
}

// AddPaste adds a new paste to BoltDB
func (b *BoltStore) AddPaste(text, lang string) (string, error) {
	id := uuid.New().String()
	paste := Paste{
		PK:       id,
		SK:       time.Now().Format(time.RFC3339),
		Language: lang,
		Text:     text,
	}

	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("pastes"))
		encoded, err := json.Marshal(paste)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(id), encoded)
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

// GetDiff retrieves a diff from BoltDB
func (b *BoltStore) GetDiff(id string) (*Diff, error) {
	var diff Diff
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("diffs"))
		v := bucket.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("diff not found")
		}
		return json.Unmarshal(v, &diff)
	})
	if err != nil {
		return nil, err
	}
	return &diff, nil
}

// AddDiff adds a new diff to BoltDB
func (b *BoltStore) AddDiff(oldText, newText string) (string, error) {
	id := uuid.New().String()
	diff := Diff{
		PK:      id,
		SK:      time.Now().Format(time.RFC3339),
		OldText: oldText,
		NewText: newText,
	}

	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("diffs"))
		encoded, err := json.Marshal(diff)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(id), encoded)
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

// Close closes the BoltDB connection
func (b *BoltStore) Close() error {
	return b.db.Close()
}

// GetPaste retrieves a paste from DynamoDB
func (d *DynamoStore) GetPaste(id string) (*Paste, error) {
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(id)},
		},
	})
	if err != nil {
		return nil, err
	}

	paste := &Paste{}
	err = dynamodbattribute.UnmarshalMap(result.Item, paste)
	if err != nil {
		return nil, err
	}

	return paste, nil
}

// AddPaste adds a new paste to DynamoDB
func (d *DynamoStore) AddPaste(text, lang string) (string, error) {
	id := uuid.New().String()
	paste := Paste{
		PK:       id,
		SK:       time.Now().Format(time.RFC3339),
		Language: lang,
		Text:     text,
	}

	av, err := dynamodbattribute.MarshalMap(paste)
	if err != nil {
		return "", err
	}

	_, err = d.svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetDiff retrieves a diff from DynamoDB
func (d *DynamoStore) GetDiff(id string) (*Diff, error) {
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(id)},
		},
	})
	if err != nil {
		return nil, err
	}

	diff := &Diff{}
	err = dynamodbattribute.UnmarshalMap(result.Item, diff)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

// AddDiff adds a new diff to DynamoDB
func (d *DynamoStore) AddDiff(oldText, newText string) (string, error) {
	id := uuid.New().String()
	diff := Diff{
		PK:      id,
		SK:      time.Now().Format(time.RFC3339),
		OldText: oldText,
		NewText: newText,
	}

	av, err := dynamodbattribute.MarshalMap(diff)
	if err != nil {
		return "", err
	}

	_, err = d.svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// Close is a no-op for DynamoDB as it doesn't require explicit closing
func (d *DynamoStore) Close() error {
	return nil
}
