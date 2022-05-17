package router

import (
	"ordermanagement/payment/middleware"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/payment/initiate", middleware.MakePayment).Methods("POST")
	return router
}
