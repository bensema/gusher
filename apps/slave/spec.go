package main

import (
	jwt "github.com/dgrijalva/jwt-go"
	"gusher/internal"
)

const (
	ConnectionEstablished   = "pusher:connection_established"
	PingEvent               = "pusher:ping"
	PongReplySucceeded      = "pusher:pong"
	Err                     = "pusher:err"
	SubscribeEvent          = "pusher:subscribe"
	SubscribeReplySucceeded = "pusher:subscription_succeeded"
	SubscribeReplyError     = "pusher:subscription_error"
	UnSubscribeEvent        = "pusher:unsubscribe"

	MultiSubscribeEvent = "pusher:multi_subscribe"

	QueryChannelEvent            = "pusher:querychannel"
	QueryChannelReplySucceeded   = "pusher:querychannel_succeeded"
	QueryChannelReplyError       = "pusher:querychannel_error"
	AddChannelEvent              = "pusher:addchannel"
	ReloadChannelEvent           = "pusher:reloadchannel"
	RemoteEvent                  = "pusher:remote"
	LoginEvent                   = "pusher:login"
	RemoteReplySucceeded         = "pusher:remote_succeeded"
	RemoteReplyError             = "pusher:remote_error"
	MultiSubscribeReplySucceeded = "pusher:multi_subscribe_succeeded"
	MultiSubscribeReplyError     = "pusher:subscription_error"
	UnSubscribeReplySucceeded    = "pusher:unsubscribe_succeeded"
	UnSubscribeReplyError        = "pusher:unsubscribe_error"
)

type BatchData struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
}

type InternalCommand struct {
	Event    string `json:"event"`
	SocketId string `json:"socket_id"`
}
type RemoteCommand struct {
	InternalCommand
	Data RemoteData `json:"data"`
}

type ChannelInfoData struct {
	InternalCommand
	Data interface{} `json:"data"`
}

type RemoteData struct {
	Remote string      `json:"remote"`
	Msg    interface{} `json:"msg"`
}
type PingCommand struct {
	InternalCommand
	Data interface{} `json:"data"`
}
type PongResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type QueryChannelResponse struct {
	InternalCommand
	Data interface{} `json:"data"`
}

type ChannelCommand struct {
	InternalCommand
	Data ChannelData `json:"data"`
}

type ChannelData struct {
	Channel interface{} `json:"channel"`
}

type ConnectionCommand struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type JwtPack struct {
	Gusher internal.Auth `json:"gusher"`
	jwt.StandardClaims
}

/*
type Auth struct {
	Channels []string        `json:"channels"`
	UserId   string          `json:"user_id"`
	AppKey   string          `json:"app_key"`
	Remotes  map[string]bool `json:"remotes"`
}
*/
type WorkerPayload struct {
	UserId   string      `json:"user_id"`
	SocketId string      `json:"socket_id"`
	Uid      string      `json:"uid"`
	AppKey   string      `json:"app_key"`
	Data     interface{} `json:"data"`
}
