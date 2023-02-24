package slogx

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func init() {
	levelVal := os.Getenv("LOG_LEVEL")
	if levelVal == "" {
		return
	}
	levelVal = strings.ToUpper(levelVal)
	var level slog.Level
	if err := json.Unmarshal([]byte(fmt.Sprintf("%q", levelVal)), &level); err != nil {
		return
	}

	ho := slog.HandlerOptions{Level: level}

	var h slog.Handler = ho.NewTextHandler(os.Stderr)
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		h = ho.NewJSONHandler(os.Stderr)
	}

	slog.SetDefault(slog.New(h))
}
