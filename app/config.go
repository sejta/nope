package app

import "time"

// Config описывает параметры HTTP-сервера.
type Config struct {
	Addr              string // например ":8080"
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration // таймаут graceful shutdown
}

// DefaultConfig возвращает безопасные дефолты.
func DefaultConfig() Config {
	return Config{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   5 * time.Second,
	}
}

func withDefaults(cfg Config) Config {
	def := DefaultConfig()
	if cfg.Addr == "" {
		cfg.Addr = def.Addr
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = def.ReadHeaderTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = def.ReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = def.WriteTimeout
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = def.IdleTimeout
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = def.ShutdownTimeout
	}
	return cfg
}
