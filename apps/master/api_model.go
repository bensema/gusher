package main

type ChannelBase struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
}

type ChannelsParams struct {
	FilterByPrefix string `json:"filter_by_prefix"`
	Info           string `json:"info"`
}

type ChannelParams struct {
	Info string `json:"info"`
}

type TriggerParams struct {
	SocketID string `json:"socket_id"`
	Info     string `json:"info"`
}

// ChannelsList represents a list of channels received by the Pusher API.
type ChannelsList struct {
	Channels map[string]ChannelListItem `json:"channels"`
}

// ChannelListItem represents an item within ChannelsList
type ChannelListItem struct {
	UserCount int `json:"user_count"`
}

// Channel represents the information about a channel from the Pusher API.
type ApiChannel struct {
	Name              string
	Occupied          bool `json:"occupied,omitempty"`
	UserCount         int  `json:"user_count,omitempty"`
	SubscriptionCount int  `json:"subscription_count,omitempty"`
}
