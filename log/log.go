package log

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
)

var (
	parentLogger *logging.Logger
	childLogger  *logging.Logger
)

func InitLog(projectName string) {
	if projectName != "" {
		client, err := logging.NewClient(context.Background(), "projects/"+projectName)
		if err != nil {
			panic(err)
		}

		err = client.Ping(context.Background())
		if err != nil {
			log.Printf("Failed to ping log client: %v", err)
			return
		}
		log.Printf("Successfully created log client: %v", client)

		parentLogger = client.Logger("request_log")
		childLogger = client.Logger("request_log_entries")
	}
}

func Log(ctx context.Context, entry logging.Entry) {
	if gc, ok := ctx.(*gin.Context); ok {
		log.Printf("gin context")
		entry.HTTPRequest = &logging.HTTPRequest{
			Request: gc.Request,
		}
	}
	log.Printf("logrequest: %+v", entry)
	if parentLogger != nil {
		parentLogger.Log(entry)
	} else {
		log.Printf("[%s] %v", entry.Severity, entry.Payload)
	}
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	Log(ctx, logging.Entry{
		Payload:  fmt.Sprintf(format, args...),
		Severity: logging.Debug,
	})
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	Log(ctx, logging.Entry{
		Payload:  fmt.Sprintf(format, args...),
		Severity: logging.Info,
	})
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	Log(ctx, logging.Entry{
		Payload:  fmt.Sprintf(format, args...),
		Severity: logging.Warning,
	})
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	Log(ctx, logging.Entry{
		Payload:  fmt.Sprintf(format, args...),
		Severity: logging.Error,
	})
}

func Criticalf(ctx context.Context, format string, args ...interface{}) {
	Log(ctx, logging.Entry{
		Payload:  fmt.Sprintf(format, args...),
		Severity: logging.Critical,
	})
}
