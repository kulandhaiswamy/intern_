package main

import (
	"fmt"
	"log"
	"net/http"
	"ordermanagement/payment/middleware"
	"ordermanagement/payment/router"
)

func main() {
	r := router.Router()
	middleware.Start_consumer_order_created()
	fmt.Println("Starting server at 8081")

	log.Fatal(http.ListenAndServe(":8081", r))

}
