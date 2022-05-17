package middleware

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"ordermanagement/order/models"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Response struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var order_rcvd models.Order_rcvd

	err := json.NewDecoder(r.Body).Decode(&order_rcvd)

	if err != nil {
		log.Fatalf("Unable to decode request body %v", err)
	}

	order_id := createOrder(order_rcvd)

	res := Response{
		Id:      order_id,
		Message: "Order created Successfully",
	}

	json.NewEncoder(w).Encode(res)
}

func GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var order_details_response models.Order_details_response

	params := mux.Vars(r)
	order_id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert string to int %v", err)
	}

	order_details_response = getOrderDetails(int64(order_id))

	json.NewEncoder(w).Encode(order_details_response)

}

func Start_consumer_order_paid() {
	go Consume_on_order_paid()
}

func Search_products(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := mux.Vars(r)
	word := params["word"]

	search_results := search_products(word)

	json.NewEncoder(w).Encode(search_results)

}

func Consume_on_order_paid() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "consumer-group-id",
		Topic:    "ORDER_PAID",
		MinBytes: 100,  // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())

		if err != nil {
			break
		}
		payment_res := new(struct {
			Order_id       int64
			Payment_status int
		})
		json.Unmarshal(m.Value, &payment_res)
		db := createConnection()
		sqlStatement := `UPDATE ORDERS SET payment_status=$1 WHERE order_id=$2`
		_, err = db.Exec(sqlStatement, payment_res.Payment_status, payment_res.Order_id)
		if err != nil {
			log.Fatalf("unable to update in orders %v", err)
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

func createConnection() *sql.DB {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Unable to open env file")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)

	}

	fmt.Println("Successfully connected")

	return db
}

func createOrder(order_rcvd models.Order_rcvd) int64 {
	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO orders(user_id,total_price,delivery_status,payment_status) VALUES($1,$2,$3,$4) RETURNING order_id`

	var order_id int64

	err := db.QueryRow(sqlStatement, order_rcvd.User_id, order_rcvd.Total_price, 0, 0).Scan(&order_id)

	if err != nil {
		log.Fatalf("Unable to create order %v", err)
	}
	fmt.Println(order_id)
	sqlStatement = `INSERT INTO order_details(order_id,item_id,quantity) VALUES($1,$2,$3)`

	for i := 0; i < len(order_rcvd.Cart); i++ {
		_, err := db.Exec(sqlStatement, order_id, order_rcvd.Cart[i].Item_id, order_rcvd.Cart[i].Quantity)
		if err != nil {
			log.Fatalf("Unable to create order %v", err)
		}
	}
	payment := new(models.Payment)

	payment.Order_id = order_id
	payment.Total_price = order_rcvd.Total_price
	payment.Amount_paid = 0
	payment.Payment_status = 0

	push_on_order_created(*payment)
	fmt.Printf("Order successfully created with id %v", order_id)

	return order_id

}

func getOrderDetails(order_id int64) models.Order_details_response {

	db := createConnection()

	defer db.Close()

	sqlStatement := `SELECT  user_id,total_price,delivery_status,payment_status FROM orders  WHERE order_id=$1`

	row := db.QueryRow(sqlStatement, order_id)
	//odr order_details_response
	var odr = new(models.Order_details_response)
	odr.Order_id = order_id
	err := row.Scan(&odr.User_id, &odr.Total_price, &odr.Delivery_status, &odr.Payment_status)
	if err != nil {
		log.Fatalf("Unable to query %v", err)
	}

	sqlStatement = `SELECT  d.item_id,i.name,i.price,i.company,d.quantity FROM order_details d,items i WHERE  d.item_id=i.item_id AND d.order_id=$1`

	rows, err := db.Query(sqlStatement, order_id)

	if err != nil {
		log.Fatalf("Unable to query %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		//ipd itemp_pack_details
		ipd := new(models.Item_pack_details)

		err := rows.Scan(&ipd.Item.Item_id, &ipd.Item.Name, &ipd.Item.Price, &ipd.Item.Company, &ipd.Quantity)
		if err != nil {
			panic(err)
		}
		odr.Items_ordered = append(odr.Items_ordered, *ipd)

	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	return *odr

}

//payment_data models.Payment
func push_on_order_created(payment models.Payment) {
	message_, _ := json.Marshal(payment)
	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "ORDER_CREATED",
		Balancer: &kafka.LeastBytes{},
	}

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(message_),
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}

}

func search_products(word string) models.Product_search_results {
	cert, _ := ioutil.ReadFile("http_ca.crt")
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Username: "elastic",
		Password: "TrcmtfjQYWEMg*h1-iY5",
		CACert:   cert,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating client %v", err)
	}
	query := map[string]interface{}{
		"_source":   []string{"PRODUCT NAME", "PRICE ( in RS)"},
		"size":      20,
		"min_score": 0.5,
		"query": map[string]interface{}{
			"prefix": map[string]interface{}{
				"PRODUCT NAME": word,
			},
		},
	}

	var mapresp map[string]interface{}
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("unable to encode %v", err)
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("product_details"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("error getting response %v", err)
	}

	defer res.Body.Close()

	//fmt.Println("hello")
	if err := json.NewDecoder(res.Body).Decode(&mapresp); err != nil {
		log.Fatalf("unable to decode %v", err)
	}
	//fmt.Println("hello")
	search_results := new(models.Product_search_results)
	for _, hit := range mapresp["hits"].(map[string]interface{})["hits"].([]interface{}) {
		product := new(models.Product_desc)
		product.Product_name = hit.(map[string]interface{})["_source"].(map[string]interface{})["PRODUCT NAME"].(string)
		price := hit.(map[string]interface{})["_source"].(map[string]interface{})["PRICE ( in RS)"].(string)
		price = strings.ReplaceAll(price, ",", "")
		product.Price, err = strconv.ParseFloat(price, 32)
		if err != nil {
			log.Fatalf("unable to convert to int %v", err)
		}
		product.Price = float64(int(product.Price*100)) / 100
		search_results.Products = append(search_results.Products, *product)
		//fmt.Println(product.Product_name, ' ', product.Price)

	}
	return *search_results

}
