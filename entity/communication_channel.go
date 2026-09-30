package entity

type CommunicationChannel string

const (
	CommunicationChannelEmail   CommunicationChannel = "EMAIL"
	CommunicationChannelPush    CommunicationChannel = "PUSH"
	CommunicationChannelMessage CommunicationChannel = "MESSAGE"
)
