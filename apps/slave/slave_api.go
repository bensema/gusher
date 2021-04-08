package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gusher/internal"
	"regexp"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

var DefaultSubHandler = func(channel string, p *internal.Payload) (err error) {
	return nil
}

type commandResponse struct {
	cmdType   string
	handler   func(string, *internal.Payload) (err error)
	msg       []byte
	data      string   // channel 名
	multiData []string //multi sub use
}

func WsConnect(c *gin.Context, rHub *internal.Hub) {

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
	d := &ConnectionCommand{}
	d.Event = ConnectionEstablished
	d.Data = map[string]interface{}{
		"socket_id":        s.SocketId(),
		"activity_timeout": 120,
	}
	_d, _ := json.Marshal(d)
	s.Send(_d)
	logger.WithField("socket_id", s.SocketId()).Info("connect")
	s.Listen(func(data []byte) (b []byte, err error) {
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
		res, err := h(appKey, s.GetAuth(), d, s.SocketId(), debug)
		if err != nil {
			logger.WithField("socket_id", s.SocketId()).WithError(err).Info("handler error")
			return
		}
		switch res.cmdType {
		case "SUB":
			s.On(res.data, res.handler)
		case "MULTISUB":
			for _, v := range res.multiData {
				s.On(v, res.handler)
			}
		case "UNSUB":
			s.Off(res.data)

		}
		return res.msg, nil
	})
	logger.WithField("socket_id", s.SocketId()).Info("disconnect")
	return

}

func UnSubscribeCommand(appkey string, auth internal.Auth, data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	channel, err := jsonparser.GetString(data, "channel")
	if err != nil {
		return
	}
	exist := false
	for _, ch := range auth.Channels {
		//新增萬用字元  如果找到這個 任何頻道皆可訂閱
		if ch == "*" {
			exist = true
			break
		}
		ech := regexp.QuoteMeta(ch)
		rch := strings.Replace(ech, `\*`, ".+", -1)
		r := regexp.MustCompile("^" + rch + "$")

		if r.MatchString(channel) {
			exist = true
			break
		}
	}
	msg = &commandResponse{
		cmdType: "UNSUB",
	}
	command := &ChannelCommand{}
	var reply []byte
	//反訂閱處理
	if exist {
		msg.data = channel
		command.Event = UnSubscribeReplySucceeded
		command.SocketId = socketId
		command.Data.Channel = channel
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
		command.Data.Channel = channel
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	}
	return
}

func SubscribeCommand(appKey string, auth internal.Auth, data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	fmt.Println(string(data))
	channel, err := jsonparser.GetString(data, "channel")
	if err != nil {
		return
	}
	msg = &commandResponse{
		handler: DefaultSubHandler,
		cmdType: "SUB",
	}
	command := &ChannelCommand{}
	//exist := false
	channelOk := false
	var reply []byte
	// Channel names should only include lower and uppercase letters, numbers and the following punctuation _ - = @ , . ;
	// As an example this is a valid channel name:
	// foo-bar_1234@=,.;
	// Public channels | Private channels private-
	r := regexp.MustCompile("^[\\w-=,.;@]+$")
	if r.MatchString(channel) {
		//exist = true
		channelOk = true
	}
	fmt.Println("channelOk:", channelOk)
	if channelOk {
		switch channelType(channel) {
		case PublicChannel:
			msg.data = channel
			command.SocketId = socketId
			command.Event = SubscribeReplySucceeded
			command.Data.Channel = channel
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
		command.Data.Channel = channel
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	}

	return
}

func MultiSubscribeCommand(appkey string, auth internal.Auth, data []byte, socketId string, debug bool) (msg *commandResponse, err error) {

	multiChannel := make([]string, 0)
	_, err = jsonparser.ArrayEach(data, func(v []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		multiChannel = append(multiChannel, string(v))

	}, "multi_channel")
	msg = &commandResponse{
		handler: DefaultSubHandler,
		cmdType: "MULTISUB",
	}
	command := &ChannelCommand{}
	var exist bool
	for _, ch := range auth.Channels {
		//新增萬用字元  如果找到這個 任何頻道皆可訂閱
		if ch == "*" {
			exist = true
			break
		}
	}
	subChannels := make([]string, 0)
	if exist {
		subChannels = multiChannel
	} else {
		isMatch := true
		for _, ch := range multiChannel {
			if !InArray(ch, auth.Channels) {
				isMatch = false
				break
			}
		}
		if isMatch {
			subChannels = multiChannel
		}
	}
	var reply []byte
	if len(subChannels) > 0 {
		msg.multiData = subChannels
		command.Event = MultiSubscribeReplySucceeded
		command.SocketId = socketId
		command.Data.Channel = subChannels
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply
	} else {

		//TODO 需重構 不讓他進入訂閱模式
		msg.cmdType = ""
		command.Event = MultiSubscribeReplyError
		command.SocketId = socketId
		command.Data.Channel = multiChannel
		reply, err = json.Marshal(command)
		if err != nil {
			return
		}
		msg.msg = reply

	}

	return
}

func PingPongCommand(appkey string, auth internal.Auth, data []byte, socketId string, debug bool) (msg *commandResponse, err error) {
	msg = &commandResponse{
		handler: DefaultSubHandler,
		cmdType: "PING",
	}

	command := &PongResponse{}
	command.Event = QueryChannelReplySucceeded
	command.SocketId = socketId
	command.Data = data
	command.Time = time.Now().Unix()

	reply, err := json.Marshal(command)
	if err != nil {
		return
	}
	msg.msg = reply
	return
}

func CommandRouter(data []byte) (fn func(appkey string, auth internal.Auth, data []byte, socketId string, debug bool) (msg *commandResponse, err error), err error) {

	val, err := jsonparser.GetString(data, "event")
	if err != nil {
		return
	}
	switch val {
	case SubscribeEvent:
		return SubscribeCommand, nil
	case MultiSubscribeEvent:
		return MultiSubscribeCommand, nil
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
