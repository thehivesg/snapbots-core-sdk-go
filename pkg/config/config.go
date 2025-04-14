package config

import "time"

type Config struct {
	BotID   string
	NATSURL string
	Timeout time.Duration
}
