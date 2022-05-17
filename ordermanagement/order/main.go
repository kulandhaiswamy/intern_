package main

import (
	"fmt"
	"log"
	"net/http"

	"ordermanagement/order/middleware"
	"ordermanagement/order/router"
)

func main() {
	r := router.Router()
	middleware.Start_consumer_order_paid()
	fmt.Println("Starting server at 8080")

	log.Fatal(http.ListenAndServe(":8080", r))

}
