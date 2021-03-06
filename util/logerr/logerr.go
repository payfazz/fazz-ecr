package logerr

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/payfazz/go-errors/v2"
)

var logFile io.Writer

func init() {
	path := os.Getenv("FAZZ_ECR_LOG_FILE")
	if path != "" {
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); err == nil {
			logFile = f
		}
	}
}

func Log(err error) error {
	if err == nil {
		return nil
	}

	if logFile != nil {
		fmt.Fprintf(logFile, "%s\n%s\n\n",
			time.Now().Format(time.RFC3339Nano),
			errors.FormatWithFilterPkgs(err,
				"main",
				"github.com/payfazz/fazz-ecr",
			),
		)
	}

	return err
}
