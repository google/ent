package main

import (
	"context"
	"io/ioutil"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/ent/log"
	"google.golang.org/api/option"
)

type LogItem struct {
	Timestamp     time.Time
	IP            string
	UserAgent     string
	RequestMethod string
	RequestURI    string
}

type LogItemGet struct {
	LogItem
	UserID   uint64
	Source   string
	Digest   []string
	Found    []string
	NotFound []string
}

type LogItemPut struct {
	LogItem
	UserID     uint64
	Source     string
	Digest     []string
	Created    []string
	NotCreated []string
}

const (
	SourceAPI = "api"
	SourceRaw = "raw"
	SourceWeb = "web"

	logsGetTable = "logs_get"
	logsPutTable = "logs_put"
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
	bigqueryDataset = bigqueryClient.Dataset(dataset)

	ensureTable(ctx, logsGetTable, LogItemGet{})
	ensureTable(ctx, logsPutTable, LogItemPut{})
}

func ensureTable(ctx context.Context, name string, st interface{}) {
	t, err := bigqueryDataset.Table(logsPutTable).Metadata(ctx)
	if err != nil {
		log.Errorf(ctx, "could not get table metadata: %v", err)
		tableSchema, err := bigquery.InferSchema(st)
		if err != nil {
			log.Errorf(ctx, "could not infer schema: %v", err)
			return
		}
		tableSchema = tableSchema.Relax()
		err = bigqueryDataset.Table(name).Create(ctx, &bigquery.TableMetadata{
			Name:   name,
			Schema: tableSchema,
		})
		if err != nil {
			log.Errorf(ctx, "could not create table: %v", err)
			return
		}
		log.Infof(ctx, "created table %q", name)
	} else {
		log.Infof(ctx, "table %q already exists: %+v", name, t)
	}
}

func LogGet(ctx context.Context, v *LogItemGet) {
	logAccess(ctx, "logs_get", v)
}

func LogPut(ctx context.Context, v *LogItemPut) {
	logAccess(ctx, "logs_put", v)
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
	log.Debugf(ctx, "logged access: %+v", v)
}
