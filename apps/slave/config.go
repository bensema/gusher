package main

import "time"

type SlaveConfig struct {
	Name              string
	LogFormatter      string
	LogInterval       time.Duration
	ScanInterval      time.Duration
	MaxMessage        int
	RedisAddr         string
	RedisDb           int
	RedisMaxIdle      int
	RedisMaxConn      int
	RedisJobAddr      string
	RedisJobDb        int
	RedisJobMaxIdle   int
	RedisJobMaxConn   int
	ApiListen         string
	ApiPrefix         string
	DecodeServiceAddr string
	Version           string
	CompileDate       string
	ExternalIp        string
	ReadBuffer        int
	WriteBuffer       int
	StartTime         time.Time
}

func (s SlaveConfig) GetStartTime() string {
	return s.StartTime.Format("2006/01/02 15:04:05")
}
