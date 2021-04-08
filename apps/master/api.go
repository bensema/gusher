package main

import (
	"crypto/rsa"
	"github.com/bensema/redisocket"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func GetAllChannelCount(c *gin.Context) {
	appKey := c.Param("app_key")

	channels, err := rsender.GetChannels(listenChannelPrefix, appKey, "*")
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get redis error"))
		return

	}
	tmp := struct {
		Count int `json:"count"`
	}{
		Count: len(channels),
	}
	b, err := json.Marshal(tmp)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)

	return

}

func GetAllChannel(c *gin.Context) {
	appKey := c.Param("app_key")
	channels, err := rsender.GetChannels(listenChannelPrefix, appKey, "*")
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get redis error"))
		return
	}
	b, err := json.Marshal(channels)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)

	return

}
func GetOnlineCountByChannel(c *gin.Context) {
	appKey := c.Param("app_key")
	channel := c.Param("channel")

	online, err := rsender.GetOnlineByChannel(listenChannelPrefix, appKey, channel)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get redis error"))
		return
	}
	tmp := struct {
		Count int `json:"count"`
	}{
		Count: len(online),
	}
	b, err := json.Marshal(tmp)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)

	return

}
func GetOnlineByChannel(c *gin.Context) {
	appKey := c.Param("app_key")
	channel := c.Param("channel")
	online, err := rsender.GetOnlineByChannel(listenChannelPrefix, appKey, channel)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get redis error"))
		return
	}
	b, err := json.Marshal(online)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)

	return

}

func GetOnlineCount(c *gin.Context) {
	appKey := c.Param("app_key")
	online, err := rsender.GetOnline(listenChannelPrefix, appKey)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get redis error"))
		return
	}
	tmp := struct {
		Count int `json:"count"`
	}{
		Count: len(online),
	}
	b, err := json.Marshal(tmp)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)
	return
}

func GetOnline(c *gin.Context) {
	app_key := c.Param("app_key")

	online, err := rsender.GetOnline(listenChannelPrefix, app_key)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))
		return
	}
	b, err := json.Marshal(online)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("json marshal error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", b)

	return

}

func PushToSocket(c *gin.Context) {

	app_key := c.Param("app_key")
	if app_key == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))
		return
	}
	socket_id := c.Param("socket_id")

	data, _ := c.GetPostForm("data")
	j := JsonCheck(data)
	rsender.PushToSid(listenChannelPrefix, app_key, socket_id, j)
	push := struct {
		SocketId string      `json:"socket_id"`
		Data     interface{} `json:"data"`
	}{
		SocketId: socket_id,
		Data:     data,
	}
	d, err := json.Marshal(push)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", []byte(d))

	return
}

func AddUserChannels(c *gin.Context) {
	app_key := c.Param("app_key")
	if app_key == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))

		return
	}
	user_id := c.Param("user_id")
	channel, _ := c.GetPostForm("data")
	if channel == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data empty error"))

		return
	}
	rsender.AddChannel(listenChannelPrefix, app_key, user_id, channel)
	push := struct {
		UserId string      `json:"user_id"`
		Data   interface{} `json:"data"`
	}{
		UserId: user_id,
		Data:   channel,
	}
	d, err := json.Marshal(push)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))

		return
	}
	a := ChannelInfoData{}
	a.Data = struct {
		Channel string `json:"channel"`
	}{
		Channel: channel,
	}
	a.Event = AddChannelEvent

	//send to user
	rsender.PushToUid(listenChannelPrefix, app_key, user_id, a)

	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)

	return

}
func ReloadUserChannels(c *gin.Context) {
	app_key := c.Param("app_key")
	if app_key == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))

		return
	}
	user_id := c.Param("user_id")
	data, _ := c.GetPostForm("data")
	channels := make([]string, 0)
	err := json.Unmarshal([]byte(data), &channels)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data unmarshal error"))

		return
	}
	rsender.ReloadChannel(listenChannelPrefix, app_key, user_id, channels)
	push := struct {
		UserId string      `json:"user_id"`
		Data   interface{} `json:"data"`
	}{
		UserId: user_id,
		Data:   data,
	}
	d, err := json.Marshal(push)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))

		return
	}
	a := ChannelInfoData{}
	a.Data = struct {
		Channels []string `json:"channels"`
	}{
		Channels: channels,
	}
	a.Event = ReloadChannelEvent

	//send to user
	rsender.PushToUid(listenChannelPrefix, app_key, user_id, a)
	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)

	return
}

func PushToUser(c *gin.Context) {
	app_key := c.Param("app_key")
	if app_key == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))
		return
	}
	user_id := c.Param("user_id")
	data, _ := c.GetPostForm("data")
	j := JsonCheck(data)
	rsender.PushToUid(listenChannelPrefix, app_key, user_id, j)
	push := struct {
		UserId string      `json:"user_id"`
		Data   interface{} `json:"data"`
	}{
		UserId: user_id,
		Data:   data,
	}
	d, err := json.Marshal(push)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))

		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)

	return
}

func PushBatchMessage(c *gin.Context) {
	app_key := c.Param("app_key")
	if app_key == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))

		return
	}
	data, _ := c.GetPostForm("batch_data")
	if data == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty batch data"))

		return
	}
	batchData := make([]BatchData, 0)
	byteData := []byte(data)
	err := json.Unmarshal(byteData, &batchData)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))

		return
	}
	l := len(batchData)
	bd := make([]redisocket.BatchData, l)
	for _, data := range batchData {
		push := struct {
			Channel string      `json:"channel"`
			Event   string      `json:"event"`
			Data    interface{} `json:"data"`
		}{
			Channel: data.Channel,
			Event:   data.Event,
			Data:    data.Data,
		}
		d, err := json.Marshal(push)
		if err != nil {
			//logger.GetRequestEntry(r).Warn(err)
			continue
		}
		b := redisocket.BatchData{
			Data:  d,
			Event: data.Channel,
		}
		bd = append(bd, b)

	}
	rsender.PushBatch(listenChannelPrefix, app_key, bd)
	response := struct {
		Total int `json:"total"`
		Cap   int `json:"cap"`
	}{
		Total: len(batchData),
		Cap:   len(byteData),
	}
	d, err := json.Marshal(response)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("response marshal error"))

		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)

	return
}

func PushMessageByPattern(c *gin.Context) {
	app_key := c.Param("app_key")
	channelPattern, _ := c.GetPostForm("channel_pattern")
	event, _ := c.GetPostForm("event")
	if app_key == "" || channelPattern == "" || event == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))

		return
	}
	re, err := regexp.Compile(channelPattern)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("channel_pattern cant regex"))

		return

	}
	chs, err := rsender.GetChannels(listenChannelPrefix, app_key, "*")
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("get channel error"))

		return

	}

	data, _ := c.GetPostForm("data")
	if data == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty data"))

		return
	}
	jsonData := JsonCheck(data)

	var match int
	for _, v := range chs {
		if re.MatchString(v) {
			match++
			push := struct {
				Channel string      `json:"channel"`
				Event   string      `json:"event"`
				Data    interface{} `json:"data"`
			}{
				Channel: v,
				Event:   event,
				Data:    jsonData,
			}
			d, err := json.Marshal(push)
			if err != nil {
				//logger.GetRequestEntry(r).Warn(err)
				continue
			}
			_, err = rsender.Push(listenChannelPrefix, app_key, v, d)
			if err != nil {
				//logger.GetRequestEntry(r).Warn(err)
				continue
			}
		}
	}
	response := struct {
		Total   int    `json:"total"`
		Pattern string `json:"pattern"`
	}{
		Total:   match,
		Pattern: channelPattern,
	}
	d, err := json.Marshal(response)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("response marshal error"))

		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)

	return

}
func PushMessage(c *gin.Context) {
	app_key := c.Param("app_key")
	channel := c.Param("channel")
	event := c.Param("event")
	if app_key == "" || channel == "" || event == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty param"))
		return
	}

	data, _ := c.GetPostForm("data")
	if data == "" {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty data"))
		return
	}
	jsonData := JsonCheck(data)

	push := struct {
		Channel string      `json:"channel"`
		Event   string      `json:"event"`
		Data    interface{} `json:"data"`
	}{
		Channel: channel,
		Event:   event,
		Data:    jsonData,
	}
	d, err := json.Marshal(push)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty data"))
		return
	}
	_, err = rsender.Push(listenChannelPrefix, app_key, channel, d)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("empty data"))
		return
	}
	c.Data(http.StatusOK, "application/json;charset=UTF-8", d)
	return
}
func DecodeJWT(c *gin.Context, key *rsa.PublicKey) {

	data, _ := c.GetPostForm("data")
	auth, err := Decode(key, data)
	if err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("data error"))
		return
	}
	if err = json.NewEncoder(c.Writer).Encode(auth); err != nil {
		c.Data(http.StatusBadRequest, "application/json;charset=UTF-8", []byte("parse error"))
		return
	}
	return

}

func Ping(c *gin.Context) {
	c.Data(http.StatusOK, "application/json;charset=UTF-8", []byte("pong"))
}
