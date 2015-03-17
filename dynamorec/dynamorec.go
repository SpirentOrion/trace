package dynamorec

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/SpirentOrion/trace"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/dynamodb"
)

type DynamoRecorder struct {
	Table *dynamodb.Table
}

func New(region string, tableName string, accessKey string, secretKey string) (*DynamoRecorder, error) {
	auth, err := aws.GetAuth(accessKey, secretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}
	server := &dynamodb.Server{
		Auth:   auth,
		Region: aws.Regions[region],
	}
	table := server.NewTable(tableName, dynamodb.PrimaryKey{KeyAttribute: dynamodb.NewNumericAttribute("SpanId", "")})
	return &DynamoRecorder{Table: table}, nil
}

func (r *DynamoRecorder) String() string {
	return fmt.Sprintf("dynamodb{%s:%s}", r.Table.Server.Region.Name, r.Table.Name)
}

func (r *DynamoRecorder) Start(s *trace.Span) error {
	hashKey := strconv.FormatInt(s.SpanId, 10)

	attrs := make([]dynamodb.Attribute, 3, 6)

	attrs[0] = dynamodb.Attribute{
		Type:  dynamodb.TYPE_NUMBER,
		Name:  "TraceId",
		Value: strconv.FormatInt(s.TraceId, 10),
	}
	attrs[1] = dynamodb.Attribute{
		Type:  dynamodb.TYPE_NUMBER,
		Name:  "ParentId",
		Value: strconv.FormatInt(s.ParentId, 10),
	}
	attrs[2] = dynamodb.Attribute{
		Type:  dynamodb.TYPE_STRING,
		Name:  "Start",
		Value: s.Start.Format(time.RFC3339Nano),
	}

	if s.Process != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "Process",
			Value: s.Process,
		})
	}

	if s.Kind != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "Kind",
			Value: s.Kind,
		})
	}

	if s.Name != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "Name",
			Value: s.Name,
		})
	}

	if len(s.Data) > 0 {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_BINARY,
			Name:  "Data",
			Value: base64.StdEncoding.EncodeToString(s.Data),
		})
	}

	_, err := r.Table.PutItem(hashKey, "", attrs)
	return err
}

func (r *DynamoRecorder) Finish(s *trace.Span) error {
	key := &dynamodb.Key{HashKey: strconv.FormatInt(s.SpanId, 10)}

	attrs := []dynamodb.Attribute{
		dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "Finish",
			Value: s.Finish.Format(time.RFC3339Nano),
		},
	}

	_, err := r.Table.UpdateAttributes(key, attrs)
	return err
}
