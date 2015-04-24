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

var _ trace.Recorder = &DynamoRecorder{}

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

func (r *DynamoRecorder) Record(s *trace.Span) error {
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
		{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "start_unix",
			Value: fmt.Sprintf("%.6f", float64(s.Start.UnixNano())/1e9),
		},
		{
			Type:  dynamodb.TYPE_STRING,
			Name:  "finish",
			Value: s.Finish.Format(time.RFC3339Nano),
		},
		{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "finish_unix",
			Value: fmt.Sprintf("%.6f", float64(s.Finish.UnixNano())/1e9),
		},
		{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "elapsed",
			Value: fmt.Sprintf("%.6f", float64(s.Finish.Sub(s.Start).Nanoseconds())/1e9),
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

	if s.DataMap != nil {
		for k, v := range s.DataMap {
			attrs = append(attrs, dynamodb.Attribute{
				Type:  dynamodb.TYPE_STRING,
				Name:  k,
				Value: fmt.Sprint(v),
			})
		}
	}

	_, err := r.Table.PutItem(traceId, spanId, attrs)
	return err
}
