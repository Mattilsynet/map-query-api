//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world map-command-api --out gen ./wit
package main

import (
	"log/slog"

	"github.com/Mattilsynet/map-query-api/pkg/nats"
	"github.com/Mattilsynet/map-query-api/pkg/subject"
	"github.com/Mattilsynet/mapis/gen/go/query/v1"
	"github.com/google/uuid"
	"go.wasmcloud.dev/component/log/wasilog"
)

const MapQueryApiName = "MapQueryApi"

var (
	conn   *nats.Conn
	js     *nats.JetStreamContext
	logger *slog.Logger
)

func init() {
	logger = wasilog.ContextLogger(MapQueryApiName)
	conn = nats.NewConn()
	var err error
	js, err = conn.Jetstream()
	if err != nil {
		logger.Error("error getting jetstreamcontext", "err", err)
		return
	}
	conn.RegisterRequestReply(MapQueryApi)
}

func MapQueryApi(msg *nats.Msg) *nats.Msg {
	logger.Info(MapQueryApiName + ": got msg: ")
	replyMsg := &nats.Msg{
		Subject: msg.Reply,
		Data:    msg.Data,
	}
	qry := &query.Query{}
	err := qry.UnmarshalVT(msg.Data)
	if err != nil {
		logger.Error("error unmarshalling msg: ", "err", err)
		replyMsg.Data = []byte(err.Error())
		return replyMsg
	}
	status := query.QueryStatus{}
	status.Id = uuid.New().String()
	qry.Status = &status
	subj := subject.NewQuerySubject(qry)
	logger.Info("MapQueryApi: subj:", "subj", subj.ToQuery())
	bytes, err := qry.MarshalVT()
	if err != nil {
		logger.Error("error marshalling msg: ", "err", err)
		replyMsg.Data = []byte(err.Error())
		return replyMsg
	}
	replyMsg.Data = bytes
	err = js.Publish(subj.ToQuery(), bytes)
	if err != nil {
		logger.Error("error publishing msg: ", "err", err)
		replyMsg.Data = []byte(err.Error())
	}
	return replyMsg
}

//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world starter-kit --out gen ./wit
func main() {}
