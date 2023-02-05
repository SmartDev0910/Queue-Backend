package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CS-PCockrill/queue/cmd/appCode/common"
	"github.com/CS-PCockrill/queue/pkg/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

// RegisterStore registers/inserts a store to database
func (app *appInjection) RegisterStore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newStore models.Store

	// Parse and decode the request body into a user struct
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	err := app.decodeJSONBody(w, r, &newStore)

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

	uid, _ := app.store.Insert(
		&newStore)
	fmt.Printf("DEBUG: UID Inserted: %v", uid)
	app.JSON(w, Map{"success": true, "error": nil})
}

// LoginStore Logs into a store
func (app *appInjection) LoginStore(w http.ResponseWriter, r *http.Request) {
	// err := r.ParseForm()
	var store models.Store
	err := app.decodeJSONBody(w, r, &store)

	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			//http.Error(w, mr.msg, mr.status)
			app.clientError(w, mr.status)
		} else {
			log.Println(err.Error())
			app.clientError(w, http.StatusInternalServerError)
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	//Check that user credentials are valid, if not send user a message,
	//then re-display the login page

	// form := forms.New(r.PostForm)
	id, err := app.store.Authenticate(store.Email, store.Password)
	//TODO: Placeholder to add ID of the current user to the session.
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		app.JSON(w, Map{"message": "unauthorized"})
		return
	}
	// Create JWT payload in order to create a token
	store, err = app.store.GetStoreFromDB(id.Hex())
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		app.JSON(w, Map{"success": false, "error": err, "store": models.Store{}})
		return
	}
	app.SetJWTCookie(w, id)
	app.JSON(w, Map{"success": true, "error": nil, "store": store})
}

func (app *appInjection) AddOneItem(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	var newItem models.Item

	// Parse and decode the request body into a user struct
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	item := app.decodeJSONBody(w, r, &newItem)

	if item != nil {
		var mr *malformedRequest
		if errors.As(item, &mr) {
			app.clientError(w, mr.status)
		} else {
			log.Println(item.Error())
			app.clientError(w, http.StatusInternalServerError)
		}
		return
	}
	uid, err := app.GetUIDFromToken(r)
	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, _ := app.store.GetStoreFromDB(uid)

	_, err = app.store.InsertOneProduct(id.Id, newItem.Name, newItem.Description, newItem.Price)
	if err != nil {
		app.JSON(w, Map{"message": "failed to insert product into Database"})
		return
	}
}

func (app *appInjection) UpdateStoreItem(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	//w.Header().Set("Content-Type", "application/json")
	err := app.decodeJSONBody(w, r, &item)
	params := mux.Vars(r)
	itemId := params["itemID"]

	sid, err := app.GetUIDFromToken(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	itemID, _ := primitive.ObjectIDFromHex(itemId)
	items, err := app.store.UpdateOneItem(sid, itemID, item.Name, item.Description, item.Price)
	if err != nil {
		app.JSON(w, Map{
			"message": "unsuccessful update.",
			"error": "this item does not match any items in your store",
		})
		return
	}
	json.NewEncoder(w).Encode(items)
}

// GetStore gets the current store logged in
func (app *appInjection) GetStore(w http.ResponseWriter, r *http.Request) {
	uid, err := app.GetUIDFromToken(r)
	if err != nil {
		//app.JSON(w, Map{"message": "unauthorized"})
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, err := app.store.GetStoreFromDB(uid)
	if err != nil {
		//app.JSON(w, Map{"success": false})
		app.clientError(w, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(id)
}

func (app *appInjection) CreateStoreLocationIndex() {
	err := app.store.CreateIndex()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("DEBUG: TESTING CreateStoreLocationIndex()")
}


func (app *appInjection) GetStores(w http.ResponseWriter, r *http.Request) {
	stores := app.store.GetStores()
	json.NewEncoder(w).Encode(stores)
}

func (app *appInjection) FilterLocation(w http.ResponseWriter, r *http.Request) {
	stores := app.store.FilterLocation()
	json.NewEncoder(w).Encode(stores)
}

func (app *appInjection) GetNearbyPoints(w http.ResponseWriter, r *http.Request) {
	var geo models.GeoJSON
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&geo)
	stores, err := app.store.GetPointsByDistance(geo, 900)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DEBUG: Stores - ", stores)
	json.NewEncoder(w).Encode(stores)
}

func (app *appInjection) GetStoreOrders(w http.ResponseWriter, r *http.Request) {
	// Store must be logged in to view its orders, thus use token associated with Request
	// Test id: 63be2dcf4cc9a5aa22df69f3
	sid, err := app.GetUIDFromToken(r)
	//fmt.Printf("UID : %s", uid)
	if err != nil {
		//app.JSON(w, http.StatusUnauthorized)
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	//filter := bson.M{"sid": bson.M{"$eq": sid}}
	orders, err := app.store.GetStoreOrders(sid)
	common.CheckError(err)
	json.NewEncoder(w).Encode(orders)
}



