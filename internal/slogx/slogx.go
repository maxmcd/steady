package slogx

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func init() {
	level := slog.LevelInfo

	levelVal := os.Getenv("LOG_LEVEL")
	if levelVal != "" {
		levelVal = strings.ToUpper(levelVal)
		if err := json.Unmarshal([]byte(fmt.Sprintf("%q", levelVal)), &level); err != nil {
			slog.Error("error unmarshal", err)
		}
	}

	ho := slog.HandlerOptions{Level: level}
	var h slog.Handler = ho.NewTextHandler(os.Stderr)
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		h = ho.NewJSONHandler(os.Stderr)
	}
	slog.SetDefault(slog.New(h))
}
