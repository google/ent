package log

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/logging"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

type Color func(format string, a ...interface{}) string

var (
	parentLogger *logging.Logger
	childLogger  *logging.Logger

	severityColor = map[logging.Severity]Color{
		logging.Debug:    color.BlueString,
		logging.Info:     color.GreenString,
		logging.Warning:  color.YellowString,
		logging.Error:    color.RedString,
		logging.Critical: color.MagentaString,
	}
)

func InitLog(projectID string) {
	if projectID != "" {
		client, err := logging.NewClient(context.Background(), "projects/"+projectID)
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
	// TODO: gRPC
	if gc, ok := ctx.(*gin.Context); ok {
		entry.HTTPRequest = &logging.HTTPRequest{
			Request: gc.Request,
		}
	}

	// Always log to stderr.
	color := severityColor[entry.Severity]
	log.Printf("[%s] %v", color("%-7s", entry.Severity), entry.Payload)

	if parentLogger != nil {
		parentLogger.Log(entry)
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
