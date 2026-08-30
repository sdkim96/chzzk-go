refactor: split chat wire format into chat_codec, verify against live API

unofficial/chat.go mixed three concerns: connection lifecycle, protocol
dispatch, and JSON framing built inline from anonymous structs. Move the wire
format into chat_codec.go so chat.go owns only the connection and no longer
imports encoding/json.

chat_codec.go replaces the per-cmd anonymous structs with one inboundFrame and
one outboundFrame. Fields that do not apply to a cmd are omitted rather than
sent empty, matching what the separate structs used to emit: a pong carries only
cmd and ver, a connect adds bdy/cid/svcid/tid, and only a send carries sid and
retry. retry is a *bool because it is sent as false, which omitempty would drop
from a plain bool.

chat.go now drives the socket directly. handshake writes cmd 100 and reads until
cmd 10100 on the raw conn before any goroutine starts, so the pumps inherit a
settled connection. connect then runs a read goroutine and, for authenticated
sessions, a write goroutine under errgroup. tid moves onto chatState, where
encodeConnect spends 1 and each encodeSend takes the next.

The write path frames outgoing messages instead of passing bytes through. The
caller no longer needs the sid, which the previous API captured but never
exposed, making authenticated send impossible to use.

transport/socket gains Read and Write for single frames. ReadLoop and WriteLoop
now go through them rather than reaching into the embedded conn.

Verified against the live API, which corrected three assumptions:

- cmd 93101 sends bdy as a bare array; only cmd 15101 wraps it in messageList.
  decodeMessageList accepts both and always returns an array. The previous
  strict decode killed the connection on the first live message. The legacy
  code's "forward bdy on unmarshal failure" was compensating for this, not
  being lenient about unknown frames.

- extras must carry extraToken, the value Chat.Token has always fetched and
  nothing ever used. Without it the server discards the frame and answers
  nothing at all, so the message silently never appears. Documented on Connect
  along with streamingChannelId being the channel ID, not the live ID.

- cmd 13101 acknowledges a send with retCode/retMsg. It was ignored, which is
  why a rejected send looked identical to a delivered one. decodeSendAck now
  surfaces a rejection as an error.

cmd 15101 was never observed in testing, so its messageList branch follows the
legacy code and stays unexercised. An undocumented cmd 94008 also appears and is
ignored.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01L8CX24RP6gJQVKxQDEj6yC
