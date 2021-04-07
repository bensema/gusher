package main

import "regexp"

type ChannelType int

const (
	_ ChannelType = iota
	PublicChannel
	PrivateChannel
	EncryptedChannel
	PresenceChannel
)

func channelType(channel string) ChannelType {
	enc := regexp.MustCompile("^private-encrypted-")
	if enc.MatchString(channel) {
		return EncryptedChannel
	}

	pri := regexp.MustCompile("^private-")
	if pri.MatchString(channel) {
		return PrivateChannel
	}

	pre := regexp.MustCompile("^presence-")
	if pre.MatchString(channel) {
		return PresenceChannel
	}

	return PublicChannel
}
