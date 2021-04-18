package main

import (
	"errors"
	"fmt"
	"github.com/bensema/redisocket"
	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	"regexp"
)

var DefaultSubHandler = func(channel string, p *redisocket.Payload) (err error) {
	return nil
}

type commandResponse struct {
	cmdType   string
	handler   func(string, *redisocket.Payload) (err error)
	msg       []byte
	data      string   // channel 名
	multiData []string // multi sub use
}

func WsConnect(c *gin.Context, rHub *redisocket.Hub) {

	appKey := c.Param("app_key")
	if appKey == "" {
		logger.Warn("app_key  is nil")
		//http.Error(w, "app_key is nil", http.StatusUnauthorized)
		return
	}

	s, err := rHub.Upgrade(c.Writer, c.Request, nil, appKey)
	if err != nil {
		logger.WithError(err).Warnf("upgrade ws connection error")
		return
	}
	defer s.Close()

	d := &BaseData{}
	d.Event = ConnectionEstablished
	d.Data, _ = json.MarshalToString(map[string]interface{}{
		"socket_id":        s.SocketId(),
		"activity_timeout": 120,
	})
	_d, _ := json.Marshal(d)
	s.Send(_d)

	logger.WithField("socket_id", s.SocketId()).Info("connect")
	s.Listen(func(data []byte) (b []byte, err error) {
		fmt.Println("data:", string(data))
		h, err := CommandRouter(data)
		if err != nil {
			logger.WithField("socket_id", s.SocketId()).WithError(err).Info("router error")
			return
		}
		d, _, _, err := jsonparser.Get(data, "data")
		if err != nil {
			logger.WithField("socket_id", s.SocketId()).WithError(err).Info("get data error")
			return
		}
		debug, err := jsonparser.GetBoolean(data, "debug")
		if err != nil {
			debug = false
		}
		res, err := h(d, s.SocketId(), debug)
		if err != nil {
			logger.WithField("socket_id", s.SocketId()).WithError(err).Info("handler error")
			return
		}
		switch res.cmdType {
		case "SUB":
			s.Sub(res.data)
			s.ActivityTime()
		case "UNSUB":
			s.UnSub(res.data)
			s.ActivityTime()
		case "PING":
			s.ActivityTime()
		}
		return res.msg, nil
	})
	logger.WithField("socket_id", s.SocketId()).Info("disconnect")
	return

}

func UnSubscribeCommand(data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	channel, err := jsonparser.GetString(data, "channel")
	if err != nil {
		return
	}
	exist := false

	msg = &commandResponse{
		cmdType: "UNSUB",
	}

	r := regexp.MustCompile("^[\\w-=,.;@]+$")
	if r.MatchString(channel) {
		exist = true
	}

	command := &BaseData{}
	var reply []byte
	//反訂閱處理
	if exist {
		msg.data = channel
		command.Event = UnSubscribeReplySucceeded
		command.SocketId = socketId
		command.Data, _ = json.MarshalToString(map[string]string{
			"channel": channel,
		})
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	} else {
		msg.data = channel

		//TODO 需重構 先不讓他進入訂閱模式
		msg.cmdType = ""
		command.Event = UnSubscribeReplyError
		command.SocketId = socketId
		command.Data, _ = json.MarshalToString(map[string]string{
			"channel": channel,
		})
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	}
	return
}

func SubscribeCommand(data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	channel, err := jsonparser.GetString(data, "channel")
	if err != nil {
		logger.Error("SubscribeCommand jsonparser err:", err)

		return
	}
	msg = &commandResponse{
		handler: DefaultSubHandler,
		cmdType: "SUB",
	}
	command := &BaseData{}
	//exist := false
	channelNameOk := false
	var reply []byte
	// Channel names should only include lower and uppercase letters, numbers and the following punctuation _ - = @ , . ;
	// As an example this is a valid channel name:
	// foo-bar_1234@=,.;
	// Public channels | Private channels private-
	r := regexp.MustCompile("^[\\w-=,.;@]+$")
	if r.MatchString(channel) {
		//exist = true
		channelNameOk = true
	}
	if channelNameOk {
		switch channelType(channel) {
		case PublicChannel:
			msg.data = channel
			command.SocketId = socketId
			command.Event = SubscribeReplySucceeded
			command.Data, _ = json.MarshalToString(map[string]string{
				"channel": channel,
			})
			reply, err = json.Marshal(command)
			if err != nil {
				return
			}
			msg.msg = reply
		case PrivateChannel:
			// todo 校验auth

		}

	} else {
		msg.cmdType = ""
		command.SocketId = socketId
		command.Event = SubscribeReplyError
		command.Data, _ = json.MarshalToString(map[string]string{
			"channel": channel,
		})
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	}

	return
}

func PingPongCommand(data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	msg = &commandResponse{
		handler: DefaultSubHandler,
		cmdType: "PING",
	}
	command := &BaseData{}
	command.Event = PongReplySucceeded
	command.Data = "{}"

	reply, err := json.Marshal(command)
	if err != nil {
		return
	}
	msg.msg = reply
	return
}

func CommandRouter(data []byte) (fn func(data []byte, socketId string, debug bool) (msg *commandResponse, err error), err error) {

	val, err := jsonparser.GetString(data, "event")
	if err != nil {
		logger.Error("CommandRouter jsonparser err:", err)
		return
	}
	switch val {
	case SubscribeEvent:
		return SubscribeCommand, nil
	case UnSubscribeEvent:
		return UnSubscribeCommand, nil
	case PingEvent:
		return PingPongCommand, nil
	default:
		err = errors.New("event errors")
		break
	}
	return
}
