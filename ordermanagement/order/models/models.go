package models

type User struct {
	User_id   int64  `json:"user_id"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Mobile_no int64  `json:"mobile_no"`
}

type Items struct {
	Item_id int64  `json:"item_id"`
	Name    string `json:"name"`
	Price   string `json:"price"`
	Company string `json:"company"`
}

type Orders struct {
	Order_id        int64 `json:"order_id"`
	User_id         int64 `json:"user_id"`
	Total_price     int32 `json:"total_price"`
	Delivery_status int   `json:"delivery_status"`
}

type Item_pack struct {
	Item_id  int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
}

type Order_rcvd struct {
	User_id     int64       `json:"user_id"`
	Cart        []Item_pack `json:"cart"`
	Total_price int         `json:"total_price"`
}

type Order_details struct {
	Order_id int64 `json:"order_id"`
	Item_id  int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
}

type Payment struct {
	Order_id       int64  `json:"order_id"`
	Total_price    int    `json:"total_price"`
	Amount_paid    int    `json:"amount_paid"`
	Payment_status int    `json:"payment_status"`
	Account_no     int64  `json:"account_no"`
	Time_stamp     string `json:"time_stamp"`
}

type Item_pack_details struct {
	Item     Items `json:"item"`
	Quantity int   `json:"quantity"`
}
type Order_details_response struct {
	Order_id        int64               `json:"order_id"`
	User_id         int64               `json:"user_id"`
	Items_ordered   []Item_pack_details `json:"items_ordered"`
	Total_price     int                 `json:"total_price"`
	Delivery_status int                 `json:"delivery_status"`
	Payment_status  int                 `json:"payment_status"`
}

type Product_desc struct {
	Product_name string  `json:"product_name"`
	Price        float64 `json:"price"`
}

type Product_search_results struct {
	Products []Product_desc `json:"products"`
}
