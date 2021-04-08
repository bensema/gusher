package main

import "time"

type MasterConfig struct {
	Name              string
	LogFormatter      string
	RedisAddr         string
	RedisDb           int
	RedisMaxIdle      int
	RedisMaxConn      int
	ApiListen         string
	ApiPrefix         string
	PublicKeyLocation string
	Version           string
	CompileDate       string
	ExternalIp        string
	StartTime         time.Time
}

func (m MasterConfig) GetStartTime() string {
	return m.StartTime.Format("2006/01/02 15:04:05")
}
