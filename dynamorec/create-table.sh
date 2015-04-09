#!/bin/sh

read -r -d '' JSON <<'EOF'
{
    "TableName": "Traces",
    "AttributeDefinitions": [
        {
            "AttributeName": "span_id",
            "AttributeType": "N"
        },
        {
            "AttributeName": "trace_id",
            "AttributeType": "N"
        }
    ],
    "KeySchema": [
        {
            "AttributeName": "span_id",
            "KeyType": "HASH"
        }
    ],
    "GlobalSecondaryIndexes": [
        {
            "IndexName": "trace_id",
            "KeySchema": [
                {
                    "AttributeName": "trace_id",
                    "KeyType": "HASH"
                }
            ],
            "Projection": {
                "ProjectionType": "KEYS_ONLY"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 5,
                "WriteCapacityUnits": 5
            }
        }
    ],
    "ProvisionedThroughput": {
        "ReadCapacityUnits": 5,
        "WriteCapacityUnits": 5
    }
}
EOF

aws dynamodb create-table --cli-input-json "$JSON"
