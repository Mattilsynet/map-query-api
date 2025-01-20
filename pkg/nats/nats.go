package nats

import (
	"errors"

	jetstreampublish "github.com/Mattilsynet/map-query-api/gen/mattilsynet/provider-jetstream-nats/jetstream-publish"
	jetstream_types "github.com/Mattilsynet/map-query-api/gen/mattilsynet/provider-jetstream-nats/types"
	"github.com/Mattilsynet/map-query-api/gen/wasmcloud/messaging/consumer"
	"github.com/Mattilsynet/map-query-api/gen/wasmcloud/messaging/handler"
	"github.com/Mattilsynet/map-query-api/gen/wasmcloud/messaging/types"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

type (
	Conn struct {
		js JetStreamContext
	}
	JetStreamContext struct{}
	Msg              struct {
		Subject string
		Reply   string
		Data    []byte
		Header  map[string][]string
	}
)

type MsgHandler func(msg *Msg)

func NewConn() *Conn {
	return &Conn{}
}

func (c *Conn) Jetstream() (*JetStreamContext, error) {
	return &c.js, nil
}

func (js *JetStreamContext) Publish(subj string, data []byte) error {
	return js.PublishMsg(&Msg{Subject: subj, Data: data})
}

func (js *JetStreamContext) PublishMsg(msg *Msg) error {
	jpMsg := jetstream_types.Msg{
		Headers: toWitNatsHeaders(msg.Header),
		Data:    cm.ToList(msg.Data),
		Subject: msg.Subject,
	}
	result := jetstreampublish.Publish(jpMsg)
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}

func toWitNatsHeaders(header map[string][]string) cm.List[jetstream_types.KeyValue] {
	keyValueList := make([]jetstream_types.KeyValue, 0)
	for k, v := range header {
		keyValueList = append(keyValueList, jetstream_types.KeyValue{
			Key:   k,
			Value: cm.ToList(v),
		})
	}
	return cm.ToList(keyValueList)
}

func FromBrokerMessageToNatsMessage(bm types.BrokerMessage) *Msg {
	if bm.ReplyTo.None() {
		return &Msg{
			Data:    bm.Body.Slice(),
			Subject: bm.Subject,
			Reply:   "",
		}
	} else {
		return &Msg{
			Data:    bm.Body.Slice(),
			Subject: bm.Subject,
			Reply:   *bm.ReplyTo.Some(),
		}
	}
}

func ToBrokenMessageFromNatsMessage(nm *Msg) types.BrokerMessage {
	if nm.Reply == "" {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.None[string](),
		}
	} else {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.Some(nm.Subject),
		}
	}
}

func (nc *Conn) Publish(msg *Msg) error {
	bm := ToBrokenMessageFromNatsMessage(msg)
	result := consumer.Publish(bm)
	if !result.IsOK() {
		return errors.New(*result.Err())
	}
	return nil
}

func (conn *Conn) RequestReply(msg *Msg, timeoutInMillis uint32) (*Msg, error) {
	bm := ToBrokenMessageFromNatsMessage(msg)
	result := consumer.Request(bm.Subject, bm.Body, timeoutInMillis)
	if result.IsOK() {
		bmReceived := result.OK()
		natsMsgReceived := FromBrokerMessageToNatsMessage(*bmReceived)
		return natsMsgReceived, nil
	} else {
		return nil, errors.New(*result.Err())
	}
}

func (conn *Conn) RegisterRequestReply(fn func(*Msg) *Msg) {
	handler.Exports.HandleMessage = func(msg types.BrokerMessage) (result cm.Result[string, struct{}, string]) {
		natsMsg := FromBrokerMessageToNatsMessage(msg)
		newMsg := fn(natsMsg)
		return consumer.Publish(ToBrokenMessageFromNatsMessage(newMsg))
	}
}

func (conn *Conn) RegisterSubscription(fn func(*Msg)) {
	handler.Exports.HandleMessage = func(msg types.BrokerMessage) (result cm.Result[string, struct{}, string]) {
		natsMsg := FromBrokerMessageToNatsMessage(msg)
		fn(natsMsg)
		return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
	}
}
