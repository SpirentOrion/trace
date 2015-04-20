#!/bin/sh

read -r -d '' JSON <<'EOF'
{
    "TableName": "Traces",
    "AttributeDefinitions": [
        {
            "AttributeName": "trace_id",
            "AttributeType": "N"
        },
        {
            "AttributeName": "span_id",
            "AttributeType": "N"
        }
    ],
    "KeySchema": [
        {
            "AttributeName": "trace_id",
            "KeyType": "HASH"
        },
        {
            "AttributeName": "span_id",
            "KeyType": "RANGE"
        }
    ],
    "ProvisionedThroughput": {
        "ReadCapacityUnits": 5,
        "WriteCapacityUnits": 5
    }
}
EOF

aws dynamodb create-table --cli-input-json "$JSON"
