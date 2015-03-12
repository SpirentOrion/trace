#!/bin/sh

aws dynamodb create-table --table-name Trace \
    --attribute-definitions AttributeName=SpanId,AttributeType=N \
    --key-schema AttributeName=SpanId,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
