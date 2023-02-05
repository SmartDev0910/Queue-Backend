package main

import (
	"encoding/json"
	"fmt"
	"github.com/CS-PCockrill/queue/cmd/appCode/common"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	//"encoding/json"
	"errors"
	"github.com/CS-PCockrill/queue/pkg/models"
	"log"
	"net/http"
)

func (app *appInjection) RegisterDriver(w http.ResponseWriter, r *http.Request) {
	var newDriver models.Driver

	// Parse and decode the request body into a user struct
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	//err := app.decodeJSONBody(w, r, &newDriver)
	err := app.decodeJSONBody(w, r, &newDriver)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			app.clientError(w, mr.status)
		} else {
			log.Println(err.Error())
			app.clientError(w, http.StatusInternalServerError)
		}
		return
	}

	_, _ = app.driver.RegisterDriver(&newDriver)
	fmt.Println("DEBUG: Inserted Driver")
	app.JSON(w, Map{"success": true, "error": nil})
}

func (app *appInjection) LoginDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var driver models.Driver
	err := app.decodeJSONBody(w, r, &driver)

	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			//http.Error(w, mr.msg, mr.status)
			app.clientError(w, mr.status)
			app.JSON(w, Map{"success": false, "error": err, "driver": models.Driver{}})
		} else {
			log.Println(err.Error())
			app.clientError(w, http.StatusInternalServerError)
			app.JSON(w, Map{"success": false, "error": err, "driver": models.Driver{}})
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	// Check that user credentials are valid, if not send user a message,
	// then re-display the login page
	fmt.Printf("DEBUG: " + driver.Email + " | " + driver.Password)
	id, err := app.driver.Authenticate(driver.Email, driver.Password)
	fmt.Printf("DRIVER: %v\n", id)
	driver, err = app.driver.GetDriver(id)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		app.JSON(w, Map{"success": false, "error": err, "driver": models.Driver{}})
		return
	}
	// Create JWT payload in order to create a token
	app.SetJWTCookie(w, id)
	app.JSON(w, Map{"success": true, "error": nil, "driver": driver})
}

func (app *appInjection) GetDriver(w http.ResponseWriter, r *http.Request) {
	uid, err := app.GetUIDFromToken(r)
	//fmt.Printf("UID : %s", uid)
	if err != nil {
		//app.JSON(w, http.StatusUnauthorized)
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, _ := primitive.ObjectIDFromHex(uid)
	driver, err := app.driver.GetDriver(id)
	json.NewEncoder(w).Encode(driver)
}

// db.driver.aggregate( drivers that are currently active )
// put those in a local variable to the function...
// Run through those values and search for the minimum distance/time from the store and return that driver document...
// Then assign that driver object to the order

//func (app *appInjection) OrderDispatch(w http.ResponseWriter, r *http.Request) {
//
//	drivers := app.driver.GetOnlineDrivers()
//	orders := app.order.FilterOrderStatus(models.Ready)
//
//	app.JSON(w, Map{"drivers": drivers, "orders": orders})
//	// FIXME: put this is a concurrent thread to run in the background
//
//}

func (app *appInjection) GetDriversNearby(w http.ResponseWriter, r *http.Request) {
	var location models.GeoQuery
	// Parse and decode the request body into a user struct
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	//err := app.decodeJSONBody(w, r, &newDriver)
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&location)
	common.CheckError(err)

	var drivers []models.Driver
	drivers, err = app.driver.GetPointsByDistance(location.Geo, location.Distance)
	common.CheckError(err)
	json.NewEncoder(w).Encode(drivers)
}

func (app *appInjection) CreateDriverLocationIndex() {
	err := app.driver.CreateIndex()
	if err != nil {
		log.Fatal(err)
	}
}

func (app *appInjection) driver_reader(conn *websocket.Conn) {
	type Accepted struct {
		DriverID string `json:"did"`
		OrderID  string	`json:"oid"`
	}

	go func() {
		for {
			//log.Println("\nTESTING ORDER READER: ", order)
			order := <-app.readyOrder
			//fmt.Printf("\nDEBUG DRIVERS: COUNT - %d\n", len(drivers))
			fmt.Println("TEST DEBUG DRIVER 1")
			orderBytes, err := json.Marshal(order)
			fmt.Println("TEST DEBUG DRIVER 2")
			if err := conn.WriteMessage(2, orderBytes); err != nil {
				log.Println("DEBUG ERROR (Driver Reader): ", err)
				return
			}
			fmt.Println("TEST DEBUG 3")
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			fmt.Println("TEST DEBUG 4")
			var accept Accepted
			err = json.Unmarshal(p, &accept)
			fmt.Println("TEST DEBUG 5")
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("\nTESTING DRIVER READER: ", accept)
		}
	}()
}

func (app *appInjection) DriverListener(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		HandshakeTimeout: time.Duration(1 * time.Hour),
	}
	upgrader.CheckOrigin = func (r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("(Driver Listener) Client successfully connected...")
	//app.driverConn = ws
	app.driver_reader(ws)
}

