package protocol

type MessageType int

const (
	MessageTypeHeatBeat MessageType = iota
	MessageTypeAuth
	MessageTypeRequest
	MessageTypeResponse
)
