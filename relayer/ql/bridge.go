package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/sithumonline/demedia-poc/core/models"
	"github.com/sithumonline/demedia-poc/core/utility/blob"
	"github.com/sithumonline/demedia-poc/core/utility/parser"
	"gorm.io/gorm"
	"log"
)

type BridgeArgs struct {
	Data []byte
}
type BridgeReply struct {
	Data []byte
}

type BridgeCall struct {
	Body   []byte
	Method string
}

type BridgeService struct {
	db *gorm.DB
}

func NewBridgeService(db *gorm.DB) *BridgeService {
	return &BridgeService{
		db: db,
	}
}

type bridge struct {
	db *gorm.DB
}

func newBridge(db *gorm.DB) *bridge {
	return &bridge{
		db: db,
	}
}

func (t *BridgeService) Ql(ctx context.Context, argType BridgeArgs, replyType *BridgeReply) error {
	b := newBridge(t.db)

	call := BridgeCall{}
	err := json.Unmarshal(argType.Data, &call)
	if err != nil {
		return err
	}

	switch call.Method {
	case "getAllItem":
		return b.getAllItem(replyType)
	case "createItem":
		return b.createItem(replyType, call.Body)
	case "fetch":
		return b.fetch(replyType, call.Body)
	case "readItem":
		return b.readItem(replyType, call.Body)
	case "file":
		return b.file(replyType, call.Body)
	default:
		return errors.New("method not found")
	}
}

func (t *bridge) getAllItem(replyType *BridgeReply) error {
	list := make([]models.Todo, 0)
	if result := t.db.Find(&list); result.Error != nil {
		log.Printf("failed to find todos: %v", result.Error)
		return result.Error
	}

	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	replyType.Data = b
	return nil
}

func (t *bridge) createItem(replyType *BridgeReply, body []byte) error {
	var d models.Todo
	err := json.Unmarshal(body, &d)
	if err != nil {
		return err
	}
	d.Id = uuid.New().String()
	if result := t.db.Create(&d); result.Error != nil {
		log.Printf("failed to create todo: %v", result.Error)
		return result.Error
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	replyType.Data = b
	return nil
}

func (t *bridge) readItem(replyType *BridgeReply, body []byte) error {
	var d models.Todo
	err := json.Unmarshal(body, &d)
	if err != nil {
		return err
	}
	if result := t.db.Where("id = ?", d.Id).First(&d); result.Error != nil {
		log.Printf("failed to find todo: %v", result.Error)
		return result.Error
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	replyType.Data = b
	return nil
}

func (t *bridge) fetch(replyType *BridgeReply, body []byte) error {
	var fetch models.Fetch
	err := json.Unmarshal(body, &fetch)
	if err != nil {
		return err
	}

	query := fetch.Query
	ok, err := parser.CheckQuery(query)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("permission denied to the target table")
	}

	rows, err := t.db.Raw(query).Rows()
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				err = json.Unmarshal(b, &v)
				if err != nil {
					log.Printf("database column error: %v", err)
				}
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}

	d, err := json.Marshal(tableData)
	if err != nil {
		return err
	}

	replyType.Data = d
	return nil
}

func (t *bridge) file(replyType *BridgeReply, body []byte) error {
	var fetch models.File
	err := json.Unmarshal(body, &fetch)
	if err != nil {
		return err
	}

	cfg := blob.AuditTrail{
		ID:        "peer_one",
		BucketURI: "s3://peer?endpoint=127.0.0.1:9000&disableSSL=true&s3ForcePathStyle=true&region=us-east-2",
	}
	blob, err := blob.NewBlobStorage(&cfg)
	defer blob.Close()
	if err != nil {
		return err
	}
	err = blob.SaveFile(fetch.Name, fetch.Data)
	if err != nil {
		return err
	}
	u, err := blob.GetFileURL(fetch.Name)
	if err != nil {
		return err
	}

	b, err := json.Marshal(models.File{Link: u, Name: fetch.Name})
	if err != nil {
		return err
	}
	replyType.Data = b
	return nil
}
