package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ordermanagement/payment/models"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Response struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
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

func MakePayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "pplication/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var payment models.Payment

	err := json.NewDecoder(r.Body).Decode(&payment)

	if err != nil {
		log.Fatalf("Unable to decode request %v", err)
	}

	time_ := makePayment(payment)

	res := Response{
		Id:      payment.Order_id,
		Message: "Successfully paid at " + time_,
	}

	json.NewEncoder(w).Encode(res)
}

func makePayment(payment models.Payment) string {

	db := createConnection()

	defer db.Close()

	var time_stamp string

	sqlStatement := `UPDATE payment SET account_no=$1,amount_paid=$2,payment_status=$3,time_stamp=CURRENT_TIMESTAMP WHERE order_id=$4 RETURNING time_stamp`

	row := db.QueryRow(sqlStatement, payment.Account_no, payment.Amount_paid, 1, payment.Order_id)
	err := row.Scan(&time_stamp)
	if err != nil {
		log.Fatalf("Unable to update %v", err)
	}
	push_on_order_paid(payment.Order_id, 1)

	return time_stamp

}

func push_on_order_paid(order_id_ int64, paymen_status_ int) {
	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "ORDER_PAID",
		Balancer: &kafka.LeastBytes{},
	}
	message_, _ := json.Marshal(struct {
		Order_id       int64
		Payment_status int
	}{
		Order_id:       order_id_,
		Payment_status: paymen_status_,
	})
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

func Start_consumer_order_created() {
	go consume_on_order_created()
}
func consume_on_order_created() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "consumer-group-id",
		Topic:    "ORDER_CREATED",
		MinBytes: 100,  // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())

		if err != nil {
			break
		}
		payment := new(models.Payment)
		json.Unmarshal(m.Value, &payment)
		db := createConnection()
		sqlStatement := `INSERT INTO payment(order_id,total_price,amount_paid,payment_status) VALUES($1,$2,$3,$4)`
		_, err = db.Exec(sqlStatement, payment.Order_id, payment.Total_price, payment.Amount_paid, payment.Payment_status)
		if err != nil {
			log.Fatalf("unable to insert in to payment %v", err)
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
