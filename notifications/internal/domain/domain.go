package domain

import (
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	desc "route256/loms/pkg/loms/v1"
)

type domain struct{}

func New() *domain {
	return &domain{}
}

func (d *domain) ReceiveOrder(data []byte) {
	var order desc.Order
	err := protojson.Unmarshal(data, &order)
	if err != nil {
		log.Println("Unmarshal order", err)
		return
	}
	log.Printf("%+v", order.String())
}
