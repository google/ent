package log

import (
	"context"
	"log"
)

func Debugf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[WARNING] "+format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func Criticalf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[CRITICAL] "+format, args...)
}
