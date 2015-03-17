#!/bin/sh

read -r -d '' JSON <<'EOF'
{
    "TableName": "Trace",
    "AttributeDefinitions": [
        {
            "AttributeName": "SpanId",
            "AttributeType": "N"
        },
        {
            "AttributeName": "TraceId",
            "AttributeType": "N"
        }
    ],
    "KeySchema": [
        {
            "AttributeName": "SpanId",
            "KeyType": "HASH"
        }
    ],
    "GlobalSecondaryIndexes": [
        {
            "IndexName": "TraceId",
            "KeySchema": [
                {
                    "AttributeName": "TraceId",
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
