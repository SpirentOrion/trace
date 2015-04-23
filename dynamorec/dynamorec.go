package dynamorec

import (
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
	table := server.NewTable(tableName, dynamodb.PrimaryKey{
		KeyAttribute:   dynamodb.NewNumericAttribute("trace_id", ""),
		RangeAttribute: dynamodb.NewNumericAttribute("span_id", ""),
	})
	return &DynamoRecorder{Table: table}, nil
}

func (r *DynamoRecorder) String() string {
	return fmt.Sprintf("dynamodb{%s:%s}", r.Table.Server.Region.Name, r.Table.Name)
}

func (r *DynamoRecorder) Start(s *trace.Span) error {
	traceId := strconv.FormatInt(s.TraceId, 10)
	spanId := strconv.FormatInt(s.SpanId, 10)

	attrs := []dynamodb.Attribute{
		{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "parent_id",
			Value: strconv.FormatInt(s.ParentId, 10),
		},
		{
			Type:  dynamodb.TYPE_STRING,
			Name:  "start",
			Value: s.Start.Format(time.RFC3339Nano),
		},
	}

	if s.Process != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "process",
			Value: s.Process,
		})
	}

	if s.Kind != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "kind",
			Value: s.Kind,
		})
	}

	if s.Name != "" {
		attrs = append(attrs, dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "name",
			Value: s.Name,
		})
	}

	_, err := r.Table.PutItem(traceId, spanId, attrs)
	return err
}

func (r *DynamoRecorder) Finish(s *trace.Span) error {
	key := &dynamodb.Key{
		HashKey:  strconv.FormatInt(s.TraceId, 10),
		RangeKey: strconv.FormatInt(s.SpanId, 10),
	}

	attrs := []dynamodb.Attribute{
		dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "finish",
			Value: s.Finish.Format(time.RFC3339Nano),
		},
	}

	if s.Data != nil {
		for k, v := range s.Data {
			attrs = append(attrs, dynamodb.Attribute{
				Type:  dynamodb.TYPE_STRING,
				Name:  k,
				Value: fmt.Sprint(v),
			})
		}
	}

	_, err := r.Table.UpdateAttributes(key, attrs)
	return err
}
