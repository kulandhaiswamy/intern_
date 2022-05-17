package models

type Payment struct {
	Order_id       int64  `json:"order_id"`
	Total_price    int    `json:"total_price"`
	Amount_paid    int    `json:"amount_paid"`
	Payment_status int    `json:"payment_status"`
	Account_no     int64  `json:"account_no"`
	Time_stamp     string `json:"time_stamp"`
}
