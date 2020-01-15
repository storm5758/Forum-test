package main

import "time"

const (
	// Database config
	Host     = "localhost"
	Port     = 5432
	User     = "user"
	Password = "password"
	DBname   = "forum"

	MaxConnIdleTime = time.Minute
	MaxConnLifetime = time.Hour
	MinConns        = 2
	MaxConns        = 4
)
