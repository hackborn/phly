package main

import (
	"crypto/rand"
	"fmt"
	"github.com/micro-go/parse"
	"math"
	"strconv"
	"time"
)

const (
	fallback_duration  = time.Duration(math.MaxInt64)
	fallback_frequency = time.Second
)

type cfg_t struct {
	id        string
	duration  time.Duration
	frequency time.Duration
}

func newCfg(args ...string) cfg_t {
	cfg := cfg_t{id: newId(4), duration: fallback_duration, frequency: fallback_frequency}
	token := parse.NewStringToken(args...)
	// Skip the app name
	token.Next()
	for cur, err := token.Next(); err == nil; cur, err = token.Next() {
		// Handle commands
		switch cur {
		case "-dur":
			n, _ := token.Next()
			cfg.duration = readDuration(n, cfg.duration)
		case "-freq":
			n, _ := token.Next()
			cfg.frequency = readDuration(n, cfg.frequency)
		}
	}
	return cfg
}

func readDuration(s string, fallback time.Duration) time.Duration {
	seconds, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds*1000) * time.Millisecond
}

func newId(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	s := fmt.Sprintf("%X", b)
	if len(s) > size {
		return s[:size]
	}
	return s
}
