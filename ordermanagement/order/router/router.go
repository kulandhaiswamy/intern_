package router

import (
	"ordermanagement/order/middleware"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/order/search/{word}", middleware.Search_products).Methods("GET")
	router.HandleFunc("/api/order/place", middleware.CreateOrder).Methods("POST")
	router.HandleFunc("/api/order/details/{id}", middleware.GetOrderDetails).Methods("GET")
	return router
}
