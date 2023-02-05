package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func (app *appInjection) myRoutes() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/user/register", app.RegisterUser).Methods("POST")
	router.HandleFunc("/user/login", app.LoginUser).Methods("POST")
	router.HandleFunc("/store/register", app.RegisterStore).Methods("POST")
	router.HandleFunc("/store/login", app.LoginStore).Methods("POST")

	router.HandleFunc("/driver/register", app.RegisterDriver).Methods("POST")
	router.HandleFunc("/driver/login", app.LoginDriver).Methods("POST")
	router.HandleFunc("/driver", app.GetDriver).Methods("GET")

	// Dummy Handlers
	router.HandleFunc("/test-status", app.TestStatus).Methods("GET")


	//TODO: Authenticate the user and app to verify the following requests
	//router.HandleFunc("/store-items/{type}", app.GetStoreItems).Methods("GET")
	// GetUsers/GetStore will return all the users and stores
	router.HandleFunc("/users", app.GetUsers).Methods("GET")
	router.HandleFunc("/stores", app.GetStores).Methods("GET")
	// Get current logged in user or store
	router.HandleFunc("/store", app.GetStore).Methods("GET")
	router.HandleFunc("/user", app.GetUser).Methods("GET")
	// Update User data, or nested document address IN User
	router.HandleFunc("/update-user", app.UpdateUser).Methods("PUT")
	router.HandleFunc("/update-address", app.UpdateAddress).Methods("PUT")
	// Logout user/store
	router.HandleFunc("/user/logout", app.Logout).Methods("DELETE")
	// Add an item (product/service) to the database
	// Update Item and AddItem but requires store id, and for that id to be the current session cookie
	router.HandleFunc("/update-item/{itemID}", app.UpdateStoreItem).Methods("PUT")
	router.HandleFunc("/add-item", app.AddOneItem).Methods("POST")

	router.HandleFunc("/order/insert", app.InsertOrder).Methods("POST")
	router.HandleFunc("/orders/{uid}", app.GetOrders).Methods("GET")

	//router.HandleFunc("/user/orders/{uid}", app.GetUserOrders)
	router.HandleFunc("/order/{orderId}", app.GetOrder).Methods("GET")

	router.HandleFunc("/order/{filter}", app.FilterStatus).Methods("GET")

	router.HandleFunc("/order/{orderID}/{status}", app.UpdateStatus).Methods("PUT")

	router.HandleFunc("/drivers-nearby", app.GetDriversNearby).Methods("POST")

	router.HandleFunc("/filter-location", app.FilterLocation).Methods("GET")
	router.HandleFunc("/nearby", app.GetNearbyPoints).Methods("POST")

	router.HandleFunc("/store/orders", app.GetStoreOrders).Methods("GET")

	router.HandleFunc("/socket/order/listen", app.ListenForOrders)
	router.HandleFunc("/socket/drivers/listen", app.DriverListener)



	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})
	handler := c.Handler(router)
	return handler
}
