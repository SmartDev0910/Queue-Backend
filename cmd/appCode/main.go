package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/CS-PCockrill/queue/pkg/models"
	"github.com/CS-PCockrill/queue/pkg/models/mongodb"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

//Application commonalities that will be accessed all over the application
type appInjection struct {
	upgrader websocket.Upgrader
	orderChan   chan models.Order
	readyOrder  chan models.Order

	user     *mongodb.UserFunctions
	store    *mongodb.StoreFunctions
	driver   *mongodb.DriverFunctions
	order    *mongodb.OrderFunctions
	errorLog *log.Logger
	infoLog  *log.Logger
	inProduction  bool
	mu sync.Mutex
}

func main() {
	addr := flag.String("addr", ":8000", "HTTP network address")
	connectionString := "mongodb+srv://queue-delivery:EmIQUAHqjAsXm9FT@cluster0.futg4.mongodb.net/queue?retryWrites=true&w=majority"
	//Make connection to the mongodb cluster
	// password := "EmIQUAHqjAsXm9FT"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		fmt.Println("MongoDB connection error!")
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	//databases, err := client.ListDatabaseNames(ctx, bson.M{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(databases)
	defer client.Disconnect(ctx)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}

	//Initialize the appInjection struct that will be passed around to rest of the application code
	//orderChan := make
	app := &appInjection{
		upgrader: upgrader,
		orderChan: make(chan models.Order),
		readyOrder: make(chan models.Order),
		user:     &mongodb.UserFunctions{CLIENT: client},
		store:    &mongodb.StoreFunctions{CLIENT: client},
		driver:   &mongodb.DriverFunctions{CLIENT: client},
		order: 	  &mongodb.OrderFunctions{CLIENT: client},
		errorLog: errLog,
		infoLog:  infoLog,
		inProduction: false,
	}

	//go func() {
	//	app.CreateStoreLocationIndex()
	//	app.CreateDriverLocationIndex()
	//	app.WatchOrders()
	//}()

	// Initialize a new http.Server struct. We set the Addr and Handler fields so
	//that the server uses the same network address and routes as before, and set
	//the ErrorLog field so that the server now uses the custom errorLog logger in
	//the event of any problems
	srv := http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.myRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,

	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errLog.Fatal(err)
}




