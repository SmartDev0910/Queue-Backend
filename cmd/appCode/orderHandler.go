package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CS-PCockrill/queue/cmd/appCode/common"
	"github.com/CS-PCockrill/queue/pkg/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (app *appInjection) InsertOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder models.Order
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&newOrder)

	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			//app.clientError(w, mr.status)
			app.JSON(w, Map{"success": false, "error": err})
		} else {
			log.Println(err.Error())
			//app.clientError(w, http.StatusInternalServerError)
			app.JSON(w, Map{"success": false, "error": err})
		}
		return
	}
	newOrder.Status = models.Received
	newOrder.Created = time.Now()

	store, err := app.store.GetStoreFromDB(newOrder.Sid)
	newOrder.StoreName = store.Name
	newOrder.Source = store.Location
	newOrder.StorePhone = store.Phone

	user, err := app.user.GetUserFromDB(newOrder.Uid)
	newOrder.UserName = user.Name
	newOrder.UserPhone = user.Phone
	newOrder.Destination = user.Location

	id, err := app.order.CreateOrder(&newOrder)
	if err != nil {
		app.JSON(w, Map{"success": false, "error": err})
		return
	}

	//order, err := app.order.GetOrder(id.Hex())
	//if err != nil {
	//	app.JSON(w, Map{"success": false, "error": err})
	//	return
	//}
	newOrder.ID = id

	app.JSON(w, Map{"success": true, "error": nil})
	app.mu.Lock()
	app.send(app.orderChan, newOrder)
	app.mu.Unlock()

}

func (app *appInjection) GetOrder(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	orderId := params["orderId"]

	//fmt.Printf("ORDERID: %v", orderId)

	order, err := app.order.GetOrder(orderId)
	if err != nil {
		app.JSON(w, Map{
			"success": false,
			"error": err,
		})
	}
	json.NewEncoder(w).Encode(order)
}

func (app *appInjection) MatchDriver(w http.ResponseWriter, r *http.Request) {

}

func (app *appInjection) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	orderId := params["orderID"]
	newStatus := params["status"]

	//newStatus := models.OrderStatus(paramStatus)
	orderID, _ := primitive.ObjectIDFromHex(orderId)
	order, err := app.order.UpdateOrderStatus(orderID, newStatus)
	//fmt.Printf("DEBUG TESTING: ORDER11 - %v", order)
	if err != nil {
		app.JSON(w, Map{
			"success": false,
			"error": err,
			"order": models.Order{},
		})
		return
	}

	//order, err := app.order.GetOrder(orderID)
	fmt.Printf("DEBUG TESTING: ORDER - %v", order)
	//ord, err := app.order.GetOrder(orderID.Hex())

	//if ord.Status == models.Ready {
	//	app.send(app.readyOrder, ord)
	//} else {
	//go
	//}
	fmt.Println("TESTING AFTER SEND")
	app.JSON(w, Map{
		"success": true,
		"error": nil,
		"order": order,
	})
	app.mu.Lock()
	app.send(app.orderChan, order)
	app.mu.Unlock()
}

func (app *appInjection) FilterStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	//paramFilter := params["filter"]
	paramFilter, _ := strconv.ParseInt(params["filter"], 10, 64)
	filter := models.OrderStatus(paramFilter)

	orders := app.order.FilterOrderStatus(filter)
	json.NewEncoder(w).Encode(orders)
}

func (app *appInjection) reader(conn *websocket.Conn) {

	go func() {

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			fmt.Println("TEST DEBUG 4")
			var order models.Order
			err = json.Unmarshal(p, &order)
			fmt.Println("TEST DEBUG 5")
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("\nTESTING ORDER READER: ", order)
			//order := <-app.readyOrder
			//fmt.Printf("\nDEBUG DRIVERS: COUNT - %d\n", len(drivers))
			fmt.Println("TEST DEBUG 1")
			orderBytes, err := json.Marshal(order)
			fmt.Println("TEST DEBUG 2")
			if order.Status == models.Ready {
				app.send(app.readyOrder, order)
			}

			if err := conn.WriteMessage(messageType, orderBytes); err != nil {
				log.Println("DEBUG ERROR (Driver Reader): ", err)
				return
			}
			fmt.Println("TEST DEBUG 3")

			//fmt.Println("app.orderchan: ", <-app.orderChan)
			//select {
			//case order := <-app.orderChan:
			//	//app.mu.Lock()
			//	//order := <-app.orderChan
			//	fmt.Printf("\nDEBUG ORDER: Listener: %v", order)
			//	//app.mu.Unlock()
			//	//if order.Status == models.Ready {
			//	//	fmt.Println("Order is in ready state!")
			//	//	app.readyOrder <- order
			//	//}
			//	//fmt.Printf("DEBUG: ORDER: %v", orderView)
			//	//log.Println(string(p))
			//	if order.Status == models.Ready {
			//		app.send(app.readyOrder, order)
			//	}
			//
			//	orderBytes, err := json.Marshal(order)
			//
			//	common.CheckError(err)
			//	if err := conn.WriteMessage(2, orderBytes); err != nil {
			//		log.Println(err)
			//		break
			//	}
			//case <-time.After(2 * time.Second):
			//	//time.Sleep(500 * time.Millisecond)
			//	continue
			//	//_, message, err := conn.ReadMessage()
			//	//if err != nil {
			//	//	fmt.Println(err)
			//	//	break
			//	//}
			//	//fmt.Println(string(message))
			//}
			//fmt.Printf("\nTESTING READER ORDER 3\n")
			//messageType, _ , _ := conn.ReadMessage()
			//if err != nil {
			//	log.Println(err)
			//	return
			//}

		}

	}()
}

func (app *appInjection) ListenForOrders(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		HandshakeTimeout: time.Duration(1 * time.Hour),
	}
	upgrader.CheckOrigin = func (r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	fmt.Printf("TESTING READER ORDER 1\n")
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("TESTING READER ORDER 2\n")

	fmt.Println("Client successfully connected...")
	app.reader(ws)
}

func (app *appInjection) WatchOrders() {
	app.order.AlertStore()
	//json.NewEncoder(w).Encode(store)
}

func (app *appInjection) GetOrders(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	//var orderViews []models.OrderView
	orders, err := app.order.GetOrders(params["uid"])

	common.CheckError(err)

	json.NewEncoder(w).Encode(orders)
}
