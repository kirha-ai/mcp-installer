package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"go.kirha.ai/mcp-installer/cmd/cli"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("failed to load .env file", slog.String("err", err.Error()))
	}

	rootCmd := cli.NewCmdRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
