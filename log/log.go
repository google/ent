package log

import (
	"context"
	"log"

	"google.golang.org/appengine/v2"
	alog "google.golang.org/appengine/v2/log"
)

var isAppEngine = appengine.IsAppEngine()

func Debugf(ctx context.Context, format string, args ...interface{}) {
	if isAppEngine {
		alog.Debugf(ctx, format, args...)
	} else {
		log.Printf("[DEBUG]"+format, args...)
	}
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	if isAppEngine {
		alog.Infof(ctx, format, args...)
	} else {
		log.Printf("[INFO]"+format, args...)
	}
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	if isAppEngine {
		alog.Warningf(ctx, format, args...)
	} else {
		log.Printf("[WARNING]"+format, args...)
	}
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	if isAppEngine {
		alog.Errorf(ctx, format, args...)
	} else {
		log.Printf("[ERROR]"+format, args...)
	}
}

func Criticalf(ctx context.Context, format string, args ...interface{}) {
	if isAppEngine {
		alog.Criticalf(ctx, format, args...)
	} else {
		log.Printf("[CRITICAL]"+format, args...)
	}
}
