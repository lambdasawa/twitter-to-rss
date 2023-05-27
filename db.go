package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

const (
	TableName = "TwitterToFeed"
)

func newDB() *dynamo.DB {
	const region = "ap-northeast-1"

	sess := session.Must(session.NewSession())

	return dynamo.New(sess, &aws.Config{Region: aws.String(region)})
}

func load(ctx context.Context, db *dynamo.DB, id string) (*Feed, error) {
	var feed Feed

	if err := db.Table(TableName).Get("ID", id).One(&feed); err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &feed, nil
}

func save(ctx context.Context, db *dynamo.DB, feed *Feed) error {
	if err := db.Table(TableName).Put(feed).RunWithContext(ctx); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}
