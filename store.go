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
	"go.uber.org/zap"
)

// DataStore is the interface for our database operations
type DataStore interface {
	GetPaste(id string) (*Paste, error)
	AddPaste(text, lang, title string) (string, error)
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
	Title    string
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
	sugar := zap.L().Sugar()

	dbType := os.Getenv("DB_TYPE")
	sugar.Infow("initializing_data_store",
		"db_type", dbType,
		"db_type_set", dbType != "",
	)

	switch dbType {
	case "bolt":
		sugar.Info("creating_bolt_store")
		return NewBoltStore()
	case "dynamo":
		sugar.Info("creating_dynamo_store")
		return NewDynamoStore()
	default:
		sugar.Infow("using_default_bolt_store", "defaulted_db_type", "bolt")
		return NewBoltStore()
	}
}

// NewBoltStore creates a new BoltStore
func NewBoltStore() (*BoltStore, error) {
	sugar := zap.L().Sugar()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pbin.db"
	}

	sugar.Infow("initializing_bolt_store",
		"db_path", dbPath,
		"db_path_from_env", os.Getenv("DB_PATH") != "",
	)

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		sugar.Errorw("failed_to_open_bolt_db",
			"db_path", dbPath,
			"error", err,
		)
		return nil, err
	}

	sugar.Info("bolt_db_opened_successfully")

	err = db.Update(func(tx *bolt.Tx) error {
		sugar.Info("creating_pastes_bucket")
		_, err := tx.CreateBucketIfNotExists([]byte("pastes"))
		if err != nil {
			sugar.Errorw("failed_to_create_pastes_bucket", "error", err)
			return fmt.Errorf("create pastes bucket: %s", err)
		}

		sugar.Info("creating_diffs_bucket")
		_, err = tx.CreateBucketIfNotExists([]byte("diffs"))
		if err != nil {
			sugar.Errorw("failed_to_create_diffs_bucket", "error", err)
			return fmt.Errorf("create diffs bucket: %s", err)
		}

		sugar.Info("bolt_buckets_created_successfully")
		return nil
	})

	if err != nil {
		sugar.Errorw("failed_to_initialize_bolt_buckets", "error", err)
		return nil, err
	}

	sugar.Info("bolt_store_initialized_successfully")
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
	sugar := zap.L().Sugar()

	sugar.Infow("attempting_to_get_paste", "id", id)

	var paste Paste
	err := b.db.View(func(tx *bolt.Tx) error {
		sugar.Info("getting_pastes_bucket_for_read")
		bucket := tx.Bucket([]byte("pastes"))
		if bucket == nil {
			sugar.Error("pastes_bucket_not_found_for_read")
			return fmt.Errorf("pastes bucket not found")
		}

		sugar.Infow("looking_up_paste_in_bolt", "id", id)
		v := bucket.Get([]byte(id))
		if v == nil {
			sugar.Warnw("paste_not_found_in_bolt", "id", id)
			return fmt.Errorf("paste not found")
		}

		sugar.Infow("paste_found_in_bolt",
			"id", id,
			"data_size", len(v),
		)

		err := json.Unmarshal(v, &paste)
		if err != nil {
			sugar.Errorw("failed_to_unmarshal_paste",
				"id", id,
				"error", err,
			)
			return err
		}

		sugar.Infow("paste_unmarshaled_successfully",
			"id", id,
			"text_length", len(paste.Text),
			"language", paste.Language,
			"title", paste.Title,
		)
		return nil
	})

	if err != nil {
		sugar.Errorw("failed_to_get_paste_from_bolt",
			"id", id,
			"error", err,
		)
		return nil, err
	}

	sugar.Infow("paste_retrieved_successfully",
		"id", id,
		"text_length", len(paste.Text),
		"language", paste.Language,
		"title", paste.Title,
	)
	return &paste, nil
}

// AddPaste adds a new paste to BoltDB
func (b *BoltStore) AddPaste(text, lang, title string) (string, error) {
	sugar := zap.L().Sugar()

	id := uuid.New().String()
	sugar.Infow("creating_paste",
		"id", id,
		"text_length", len(text),
		"language", lang,
		"title", title,
		"has_text", text != "",
	)

	paste := Paste{
		PK:       id,
		SK:       time.Now().Format(time.RFC3339),
		Language: lang,
		Text:     text,
		Title:    title,
	}

	sugar.Info("starting_bolt_transaction")
	err := b.db.Update(func(tx *bolt.Tx) error {
		sugar.Info("getting_pastes_bucket")
		bucket := tx.Bucket([]byte("pastes"))
		if bucket == nil {
			sugar.Error("pastes_bucket_not_found")
			return fmt.Errorf("pastes bucket not found")
		}

		sugar.Info("marshaling_paste_to_json")
		encoded, err := json.Marshal(paste)
		if err != nil {
			sugar.Errorw("failed_to_marshal_paste", "error", err)
			return err
		}

		sugar.Infow("writing_paste_to_bolt",
			"id", id,
			"encoded_size", len(encoded),
		)

		err = bucket.Put([]byte(id), encoded)
		if err != nil {
			sugar.Errorw("failed_to_write_paste_to_bolt",
				"id", id,
				"error", err,
			)
			return err
		}

		sugar.Infow("paste_written_successfully",
			"id", id,
			"encoded_size", len(encoded),
		)
		return nil
	})

	if err != nil {
		sugar.Errorw("bolt_transaction_failed",
			"id", id,
			"error", err,
		)
		return "", err
	}

	sugar.Infow("paste_added_successfully",
		"id", id,
		"text_length", len(text),
		"language", lang,
		"title", title,
	)
	return id, nil
}

// GetDiff retrieves a diff from BoltDB
func (b *BoltStore) GetDiff(id string) (*Diff, error) {
	sugar := zap.L().Sugar()

	sugar.Infow("attempting_to_get_diff", "id", id)

	var diff Diff
	err := b.db.View(func(tx *bolt.Tx) error {
		sugar.Info("getting_diffs_bucket_for_read")
		bucket := tx.Bucket([]byte("diffs"))
		if bucket == nil {
			sugar.Error("diffs_bucket_not_found_for_read")
			return fmt.Errorf("diffs bucket not found")
		}

		sugar.Infow("looking_up_diff_in_bolt", "id", id)
		v := bucket.Get([]byte(id))
		if v == nil {
			sugar.Warnw("diff_not_found_in_bolt", "id", id)
			return fmt.Errorf("diff not found")
		}

		sugar.Infow("diff_found_in_bolt",
			"id", id,
			"data_size", len(v),
		)

		err := json.Unmarshal(v, &diff)
		if err != nil {
			sugar.Errorw("failed_to_unmarshal_diff",
				"id", id,
				"error", err,
			)
			return err
		}

		sugar.Infow("diff_unmarshaled_successfully",
			"id", id,
			"old_text_length", len(diff.OldText),
			"new_text_length", len(diff.NewText),
		)
		return nil
	})

	if err != nil {
		sugar.Errorw("failed_to_get_diff_from_bolt",
			"id", id,
			"error", err,
		)
		return nil, err
	}

	sugar.Infow("diff_retrieved_successfully",
		"id", id,
		"old_text_length", len(diff.OldText),
		"new_text_length", len(diff.NewText),
	)
	return &diff, nil
}

// AddDiff adds a new diff to BoltDB
func (b *BoltStore) AddDiff(oldText, newText string) (string, error) {
	sugar := zap.L().Sugar()

	id := uuid.New().String()
	sugar.Infow("creating_diff",
		"id", id,
		"old_text_length", len(oldText),
		"new_text_length", len(newText),
		"has_old_text", oldText != "",
		"has_new_text", newText != "",
	)

	diff := Diff{
		PK:      id,
		SK:      time.Now().Format(time.RFC3339),
		OldText: oldText,
		NewText: newText,
	}

	sugar.Info("starting_bolt_transaction_for_diff")
	err := b.db.Update(func(tx *bolt.Tx) error {
		sugar.Info("getting_diffs_bucket")
		bucket := tx.Bucket([]byte("diffs"))
		if bucket == nil {
			sugar.Error("diffs_bucket_not_found")
			return fmt.Errorf("diffs bucket not found")
		}

		sugar.Info("marshaling_diff_to_json")
		encoded, err := json.Marshal(diff)
		if err != nil {
			sugar.Errorw("failed_to_marshal_diff", "error", err)
			return err
		}

		sugar.Infow("writing_diff_to_bolt",
			"id", id,
			"encoded_size", len(encoded),
		)

		err = bucket.Put([]byte(id), encoded)
		if err != nil {
			sugar.Errorw("failed_to_write_diff_to_bolt",
				"id", id,
				"error", err,
			)
			return err
		}

		sugar.Infow("diff_written_successfully",
			"id", id,
			"encoded_size", len(encoded),
		)
		return nil
	})

	if err != nil {
		sugar.Errorw("bolt_transaction_failed_for_diff",
			"id", id,
			"error", err,
		)
		return "", err
	}

	sugar.Infow("diff_added_successfully",
		"id", id,
		"old_text_length", len(oldText),
		"new_text_length", len(newText),
	)
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
func (d *DynamoStore) AddPaste(text, lang, title string) (string, error) {
	sugar := zap.L().Sugar()

	id := uuid.New().String()
	sugar.Infow("creating_paste_in_dynamo",
		"id", id,
		"text_length", len(text),
		"language", lang,
		"title", title,
		"table_name", d.tableName,
		"has_text", text != "",
	)

	paste := Paste{
		PK:       id,
		SK:       time.Now().Format(time.RFC3339),
		Language: lang,
		Text:     text,
		Title:    title,
	}

	sugar.Info("marshaling_paste_for_dynamo")
	av, err := dynamodbattribute.MarshalMap(paste)
	if err != nil {
		sugar.Errorw("failed_to_marshal_paste_for_dynamo", "error", err)
		return "", err
	}

	sugar.Infow("writing_paste_to_dynamo",
		"id", id,
		"table_name", d.tableName,
		"item_size", len(av),
	)

	_, err = d.svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		sugar.Errorw("failed_to_write_paste_to_dynamo",
			"id", id,
			"table_name", d.tableName,
			"error", err,
		)
		return "", err
	}

	sugar.Infow("paste_added_successfully_to_dynamo",
		"id", id,
		"text_length", len(text),
		"language", lang,
		"title", title,
		"table_name", d.tableName,
	)
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
