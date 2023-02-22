package main

import (
	"log"
	"net/http"
	"route256/libs/srvwrapper"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	"route256/loms/internal/handlers/cancelorder"
	"route256/loms/internal/handlers/createorder"
	"route256/loms/internal/handlers/listorder"
	"route256/loms/internal/handlers/orderpayed"
	"route256/loms/internal/handlers/stocks"

	"github.com/gorilla/mux"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal("config init", err)
	}

	domain := domain.New()

	stocksHandler := stocks.New(domain)
	createOrderHandler := createorder.New(domain)
	listOrderHandler := listorder.New(domain)
	orderPayedHandler := orderpayed.New(domain)
	cancelOrderHandler := cancelorder.New(domain)

	r := mux.NewRouter()
	r.Handle("/stocks", srvwrapper.New(stocksHandler.Handle)).Methods("POST")

	r.Handle("/createOrder", srvwrapper.New(createOrderHandler.Handle)).Methods("POST")

	r.Handle("/listOrder", srvwrapper.New(listOrderHandler.Handle)).Methods("POST")

	r.Handle("/orderPayed", srvwrapper.New(orderPayedHandler.Handle)).Methods("POST")

	r.Handle("/cancelOrder", srvwrapper.New(cancelOrderHandler.Handle)).Methods("POST")

	log.Println("listening http at", config.ConfigData.Port)
	err = http.ListenAndServe(config.ConfigData.Port, r)
	log.Fatal("cannot listen http", err)
}
