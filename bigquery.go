package main

import (
	"context"
	"io/ioutil"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/ent/log"
	"google.golang.org/api/option"
)

type LogItemGet struct {
	Timestamp     time.Time
	IP            string
	UserAgent     string
	RequestMethod string
	RequestURI    string
	APIKey        string
	Digest        []string
	Found         []string
	NotFound      []string
	Source        string
}

type LogItemPut struct {
	Timestamp     time.Time
	IP            string
	UserAgent     string
	RequestMethod string
	RequestURI    string
	APIKey        string
	Digest        []string
	Created       []string
	NotCreated    []string
	Source        string
}

const (
	SourceAPI = "api"
	SourceRaw = "raw"
	SourceWeb = "web"
)

var bigqueryDataset *bigquery.Dataset

func InitBigquery(ctx context.Context, dataset string) {
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
	bigqueryDataset = bigqueryClient.Dataset(dataset) //.Table(table)
}

func LogGet(ctx context.Context, v *LogItemGet) {
	logAccess(ctx, "access_logs_get", v)
}

func LogPut(ctx context.Context, v *LogItemPut) {
	logAccess(ctx, "access_logs_put", v)
}

func logAccess(ctx context.Context, table string, v interface{}) {
	if bigqueryDataset == nil {
		return
	}
	t := bigqueryDataset.Table(table)
	log.Debugf(ctx, "logging access: %+v", v)
	err := t.Inserter().Put(ctx, v)
	if err != nil {
		log.Errorf(ctx, "could not insert into bigquery: %v", err)
		return
	}
}
