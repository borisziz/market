package main

import (
	"log"
	"net/http"
	"route256/checkout/internal/clients/loms"
	"route256/checkout/internal/clients/productservice"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	"route256/checkout/internal/handlers/addtocart"
	"route256/checkout/internal/handlers/deletefromcart"
	"route256/checkout/internal/handlers/listcart"
	"route256/checkout/internal/handlers/purchase"
	"route256/libs/srvwrapper"

	"github.com/gorilla/mux"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal("config init", err)
	}

	lomsClient := loms.New(config.ConfigData.Services.Loms)
	productsServiceClient := productservice.New(config.ConfigData.Token, config.ConfigData.Services.Products)

	businessLogic := domain.New(lomsClient, productsServiceClient)

	addToCartHandler := addtocart.New(businessLogic)
	deleteFromCartHandler := deletefromcart.New(businessLogic)
	listCartHandler := listcart.New(businessLogic)
	purchaseHandler := purchase.New(businessLogic)

	r := mux.NewRouter()
	r.Handle("/addToCart", srvwrapper.New(addToCartHandler.Handle)).Methods("POST")

	r.Handle("/deleteFromCart", srvwrapper.New(deleteFromCartHandler.Handle)).Methods("POST")

	r.Handle("/listCart", srvwrapper.New(listCartHandler.Handle)).Methods("POST")

	r.Handle("/purchase", srvwrapper.New(purchaseHandler.Handle)).Methods("POST")

	log.Println("listening http at", config.ConfigData.Port)
	err = http.ListenAndServe(config.ConfigData.Port, r)
	log.Fatal("cannot listen http", err)
}
