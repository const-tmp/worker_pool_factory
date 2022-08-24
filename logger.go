package worker_pool_factory

import (
	"log"
	"os"
)

func DefaultLogger(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, log.Lshortfile|log.LstdFlags|log.Lmsgprefix)
}
