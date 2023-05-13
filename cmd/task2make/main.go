package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/vkd/gowalker"
	"github.com/vkd/gowalker/config"
	"gopkg.in/yaml.v3"

	"github.com/vkd/task2make"
)

type Config struct {
	TaskFilepath   string `flag:"taskfile"`
	OutputFilepath string `flag:"output"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := mainCtx(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func mainCtx(ctx context.Context) error {
	var cfg Config
	err := config.Default(&cfg)
	if err != nil {
		if errors.Is(err, gowalker.ErrPrintHelp) {
			return nil
		}
		return fmt.Errorf("parse config: %w", err)
	}

	tsBs, err := os.ReadFile(cfg.TaskFilepath)
	if err != nil {
		return fmt.Errorf("read Taskfile file (%q): %w", cfg.TaskFilepath, err)
	}

	var taskfile task2make.Taskfile
	err = yaml.NewDecoder(bytes.NewReader(tsBs)).Decode(&taskfile)
	if err != nil {
		return fmt.Errorf("decode yaml of Taskfile: %w", err)
	}

	mk, err := os.OpenFile(cfg.OutputFilepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open output file (%q): %w", cfg.OutputFilepath, err)
	}
	defer mk.Close()

	err = taskfile.WriteMakefile(mk)
	if err != nil {
		return fmt.Errorf("write makefile: %w", err)
	}

	return nil
}
