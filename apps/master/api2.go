package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func apiEvents(c *gin.Context) {
	m := struct {
		Channels []string    `json:"channels"`
		Name     string      `json:"name"`
		Data     interface{} `json:"data"`
	}{}
	_bf, err := c.GetRawData()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err = json.Unmarshal(_bf, &m)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if len(m.Channels) < 0 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	event := m.Name
	data, _ := json.MarshalToString(m.Data)
	appKey := c.Param("app_key")
	if appKey == "" || event == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	for _, channel := range m.Channels {
		if channel == "" {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		push := ChannelBase{
			Channel: channel,
			Event:   event,
			Data:    data,
		}
		d, err := json.Marshal(push)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		_, err = rsender.Push(listenChannelPrefix, appKey, channel, d)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{})
	return
}

func apiBatchEvents(c *gin.Context) {
	m := struct {
		Batch []ChannelBase `json:"batch"`
	}{}

	_bf, err := c.GetRawData()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err = json.Unmarshal(_bf, &m)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	appKey := c.Param("app_key")
	for _, e := range m.Batch {
		d, err := json.MarshalToString(e.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		_, err = rsender.Push(listenChannelPrefix, appKey, e.Channel, []byte(d))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
	}

}

func apiChannels(c *gin.Context) {
	appKey := c.Param("app_key")
	channels, err := rsender.GetChannels(listenChannelPrefix, appKey, "*")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	//b, err := json.Marshal(channels)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{})
	//	return
	//}
	reply := struct {
		Channels map[string]interface{} `json:"channels"`
	}{}
	reply.Channels = make(map[string]interface{})
	for _, channel := range channels {
		reply.Channels[channel] = make(map[string]interface{})
	}
	c.JSON(http.StatusOK, reply)

	return

}

func apiChannel(c *gin.Context) {
	// todo get channel info
	//appKey := c.Param("app_key")
	channel := c.Param("channel")

	reply := struct {
		Channel string `json:"channel"`
	}{}
	reply.Channel = channel
	c.JSON(http.StatusOK, reply)

	return

}

func channelUsers(c *gin.Context) {
	// todo for presence-channel

}
