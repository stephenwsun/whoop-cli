package main

import (
	"context"
	"os"

	"github.com/stephensun/whoop-cli/internal/cli"
)

func main() {
	cli.Run(context.Background(), os.Args[1:])
}
