package unofficial

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// This file owns the Chzzk chat wire format: how frames are built and taken
// apart. chat.go owns the connection lifecycle and never touches JSON directly.
//
// Every frame shares the same envelope. cmd identifies the message, bdy carries
// the payload, and tid correlates a request with its response. The server
// assigns sid when it acks the connect, and every later outgoing frame must
// echo it back.

// inboundFrame is the envelope common to everything the server sends.
type inboundFrame struct {
	Bdy json.RawMessage `json:"bdy"`
	Cmd int             `json:"cmd"`
}

// connectedBody is the bdy of a cmd 10100 acknowledgement.
type connectedBody struct {
	Sid string `json:"sid"`
}

// chatBody is the bdy of cmd 15101, which batches its messages under a
// messageList key. cmd 93101 sends the same batch as a bare array instead, so
// decodeMessageList accepts both.
type chatBody struct {
	MessageList json.RawMessage `json:"messageList"`
}

// outboundFrame is the envelope for everything the client sends. Fields that do
// not apply to a given cmd are omitted rather than sent empty, matching the
// per-cmd structs this replaced: a pong carries only cmd and ver, a connect adds
// bdy/cid/svcid/tid, and only a send carries sid and retry.
//
// Retry is a pointer because it is sent as false on cmd 3101, which a plain bool
// with omitempty would drop.
type outboundFrame struct {
	Bdy   any    `json:"bdy,omitempty"`
	Cid   string `json:"cid,omitempty"`
	Cmd   int    `json:"cmd"`
	Retry *bool  `json:"retry,omitempty"`
	Sid   string `json:"sid,omitempty"`
	Svcid string `json:"svcid,omitempty"`
	Tid   int    `json:"tid,omitempty"`
	Ver   string `json:"ver"`
}

// connectBody is the bdy of a cmd 100 request. Uid is nil for anonymous
// read-only sessions and the user ID hash for authenticated ones.
type connectBody struct {
	Uid     any    `json:"uid"`
	DevType int    `json:"devType"`
	AccTkn  string `json:"accTkn"`
	Auth    string `json:"auth"`
}

// decodeFrame splits a raw server frame into its cmd and bdy.
func decodeFrame(data []byte) (int, json.RawMessage, error) {
	var f inboundFrame
	if err := json.Unmarshal(data, &f); err != nil {
		return 0, nil, fmt.Errorf("chat: decode frame: %w", err)
	}
	return f.Cmd, f.Bdy, nil
}

// decodeSendAck reports whether a cmd 13101 acknowledgement accepted the message
// it answers, returning an error describing the rejection otherwise.
//
// retCode 0 with retMsg SUCCESS means the message was posted. A rejected send is
// only ever reported here: the server drops the frame silently otherwise, so a
// caller that ignores this command cannot tell a delivered message from a
// discarded one.
//
// tid comes back as a string even though it is sent as a number, so it is not
// decoded here.
func decodeSendAck(data []byte) error {
	var ack struct {
		RetCode *int   `json:"retCode"`
		RetMsg  string `json:"retMsg"`
	}
	if err := json.Unmarshal(data, &ack); err != nil {
		return fmt.Errorf("chat: decode send ack: %w", err)
	}
	if ack.RetCode == nil || *ack.RetCode == 0 {
		return nil
	}
	return fmt.Errorf("chat: server rejected the message: retCode=%d retMsg=%q", *ack.RetCode, ack.RetMsg)
}

// decodeConnected extracts the session ID from a cmd 10100 body.
func decodeConnected(bdy json.RawMessage) (string, error) {
	var b connectedBody
	if err := json.Unmarshal(bdy, &b); err != nil {
		return "", fmt.Errorf("chat: decode connected: %w", err)
	}
	if b.Sid == "" {
		return "", fmt.Errorf("chat: server acked the connect without a sid")
	}
	return b.Sid, nil
}

// decodeMessageList extracts the batched messages from a cmd 93101 or 15101
// body, always returning them as a JSON array so callers see one shape.
//
// The two commands disagree on the envelope: 93101 sends a bare array of
// messages, while 15101 wraps the same array in a messageList object. Both are
// accepted here.
//
// It returns nil when the batch is empty, which callers should skip rather than
// forward.
func decodeMessageList(bdy json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(bdy)
	if len(trimmed) == 0 {
		return nil, nil
	}

	list := trimmed
	if trimmed[0] != '[' {
		var b chatBody
		if err := json.Unmarshal(trimmed, &b); err != nil {
			return nil, fmt.Errorf("chat: decode message list: %w", err)
		}
		list = bytes.TrimSpace(b.MessageList)
		if len(list) == 0 {
			return nil, nil
		}
	}

	var elems []json.RawMessage
	if err := json.Unmarshal(list, &elems); err != nil {
		return nil, fmt.Errorf("chat: decode message list: %w", err)
	}
	if len(elems) == 0 {
		return nil, nil
	}
	return list, nil
}

// encodeConnect builds the cmd 100 request that opens a chat session. It spends
// tid 1, so outgoing messages start at 2.
//
// uid is nil and auth is READ for an anonymous session; an authenticated one
// sends the user ID hash and SEND.
func (st *chatState) encodeConnect(uid any, auth string) ([]byte, error) {
	st.IncrementTid()
	return json.Marshal(outboundFrame{
		Bdy: connectBody{
			Uid:     uid,
			DevType: 2001,
			AccTkn:  st.token.AccessToken,
			Auth:    auth,
		},
		Cid:   st.liveID,
		Cmd:   cmdConnect,
		Svcid: "game",
		Tid:   st.tid,
		Ver:   "2",
	})
}

// encodePong builds the cmd 10000 reply to a server ping. It carries no body,
// no cid and no tid.
func encodePong() ([]byte, error) {
	return json.Marshal(outboundFrame{Cmd: cmdPong, Ver: "2"})
}

// encodeSend wraps a message body in the cmd 3101 envelope, stamping it with the
// sid from the handshake and the next tid.
//
// Only the envelope is built here: body is the caller's bdy object, passed
// through verbatim. The server expects it to carry msg, msgTypeCode, msgTime and
// extras, where extras is itself a JSON *string* holding chatType, emojis, osType
// and streamingChannelId. streamingChannelId is the channel ID, which this
// package never receives — Connect takes only the live ID — so the body stays the
// caller's responsibility.
//
// tid is only touched by the handshake, which runs before any goroutine starts,
// and by the single write goroutine, so it needs no synchronisation.
func (st *chatState) encodeSend(body []byte) ([]byte, error) {
	if st.sid == "" {
		return nil, fmt.Errorf("chat: cannot send before the handshake assigns a sid")
	}
	if !json.Valid(body) {
		return nil, fmt.Errorf("chat: send body must be valid JSON")
	}

	st.IncrementTid()
	retry := false
	return json.Marshal(outboundFrame{
		Bdy:   json.RawMessage(body),
		Cid:   st.liveID,
		Cmd:   cmdSend,
		Retry: &retry,
		Sid:   st.sid,
		Svcid: "game",
		Tid:   st.tid,
		Ver:   "2",
	})
}
