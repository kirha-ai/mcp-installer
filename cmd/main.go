package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kirha-ai/logger"
	"github.com/kirha-ai/mcp-installer/cmd/cli"
)

func main() {
	logger.Configure(
		logger.WithApplication("mcp-installer"),
	)

	log := logger.New("main")

	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warn("failed to load .env file", logger.Error(err))
	}

	rootCmd := cli.NewCmdRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
