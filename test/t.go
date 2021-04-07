package main

import (
	"github.com/pusher/pusher-http-go/v5"
	"time"
)

func main() {
	pusherClient := pusher.Client{
		AppID:   "785449",
		Key:     "c87c298e4bad3867c980",
		Secret:  "4b423c00f0289f3f37da",
		Cluster: "ap3",
		Secure:  true,
	}

	//pusherClient,_ := pusher.ClientFromURL("http://c87c298e4bad3867c980:4b423c00f0289f3f37da@api.bensemasss.com/apps/785449")

	data := map[string]string{"message": "hello world"}
	pusherClient.Trigger("my-channel", "my-event", data)

	for {
		time.Sleep(1 * time.Second)
		pusherClient.Trigger("my-channel", "my-event", data)
	}

	//attributes := "user_count,subscription_count"
	//params := pusher.ChannelParams{Info: &attributes}
	//fmt.Println(pusherClient.Channel("my-channel",params))
}
