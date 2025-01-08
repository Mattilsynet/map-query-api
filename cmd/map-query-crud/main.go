package main

import (
	"log"
	"time"

	metav1 "github.com/Mattilsynet/map-types/gen/go/meta/v1"
	"github.com/Mattilsynet/map-types/gen/go/query/v1"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Error connecting to nats: %v", err)
	}
	qry := query.Query{}
	qry.Spec = &query.QuerySpec{Action: "GET", Session: "123", Type: &metav1.TypeMeta{Kind: "ManagedEnvironment", ApiVersion: "v1"}}
	bytes, err := qry.MarshalVT()
	if err != nil {
		log.Fatalf("Error marshalling query: %v", err)
	}
	res, err := nc.Request("map.get", bytes, 10*time.Second)
	if err != nil {
		log.Fatalf("Error sending command: %v", err)
	}
	log.Println("response: ", string(res.Data))
}
