package main

import (
	"context"
	"io/ioutil"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/ent/log"
	"google.golang.org/api/option"
)

type AccessItem struct {
	Timestamp2    time.Time
	IP            string
	UserAgent     string
	RequestMethod string
	RequestURI    string
	Operation     string
	APIKey        string
	Requested     []string
	Found         []string
	NotFound      []string
	Source        string
}

const (
	OperationGet = "get"
	OperationPut = "put"

	SourceAPI = "api"
	SourceWeb = "web"
)

var bigqueryTable *bigquery.Table

func InitBigquery(ctx context.Context) {
	opts := []option.ClientOption{}
	c, _ := ioutil.ReadFile("./credentials.json")
	if len(c) > 0 {
		log.Infof(ctx, "using credentials file")
		opts = append(opts, option.WithCredentialsJSON(c))
	} else {
		log.Infof(ctx, "using application default credentials")
	}
	bigqueryClient, err := bigquery.NewClient(ctx, bigquery.DetectProjectID, opts...)
	if err != nil {
		log.Errorf(ctx, "could not create bigquery client: %v", err)
	}
	bigqueryTable = bigqueryClient.Dataset("access_logs").Table("logs")
}

func LogAccess(ctx context.Context, v AccessItem) {
	log.Debugf(ctx, "logging access: %+v", v)
	if bigqueryTable == nil {
		return
	}
	err := bigqueryTable.Inserter().Put(ctx, v)
	if err != nil {
		log.Errorf(ctx, "could not insert into bigquery: %v", err)
		return
	}
}
