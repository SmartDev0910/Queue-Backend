package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

type OrderStatus int
const (
	Received OrderStatus = iota
	Preparing
	Ready
	Queued
	Delivered
)

type User struct {
	ID primitive.ObjectID    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email          string    `json:"email,omitempty" validate:"required,email"`
	Phone 		   string    `json:"phone,omitempty"`
	Password 	   string    `json:"password,omitempty" validate:"required"`
	HashedPassword []byte 	 `json:"-"`
	Created        time.Time	 `json:"-"`
	Active         bool      `json:"-"`
	Location       GeoJSON	 `json:"location,omitempty" bson:"location"`
	Address        []Address   `json:"address,omitempty"`
	Orders         []Order   `json:"orders,omitempty" bson:"orders"`
}

type Store struct {
	Id          	primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Owner			string 			   `json:"owner"`
	Email 			string 			   `json:"email"`
	Name        	string             `json:"name"`
	Phone 			string 			   `json:"phone"`
	Password 		string 			   `json:"password,omitempty"`
	HashedPassword  []byte 		       `json:"-"`
	Location		GeoJSON				`json:"location"`
	Address     	Address            `json:"address,omitempty" bson:"address,omitempty"`
	Inventory       []Item          `json:"inventory" bson:"inventory"`
}

type GeoJSON struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Coordinates []float64 `json:"coordinates,omitempty" bson:"coordinates,omitempty"`
}

type GeoQuery struct {
	Distance int `json:"distance" bson:"distance"`
	Geo      GeoJSON `json:"geo" bson:"geo"`
}

type Point struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	Title string		`json:"title"`
	Location GeoJSON    `json:"location"`
}

//Address data type for the address object
type Address struct {
	Street string             `json:"street,omitempty" bson:"street,omitempty"`
	Suite string			   `json:"suite,omitempty" bson:"suite,omitempty"`
	City    string             `json:"city,omitempty" bson:"city,omitempty"`
	State   string             `json:"state,omitempty" bson:"state,omitempty"`
	Country string			   `json:"country,omitempty" bson:"country,omitempty"`
	Zip     string             `json:"zip,omitempty" bson:"zip,omitempty"`
}

type Item struct {
	Id 			primitive.ObjectID	   			`json:"id" bson:"_id,omitempty"`
	StoreId       string   			 `json:"sid" bson:"sid,omitempty"`
	Name          string   			 `json:"name" bson:"name"`
	Price         float64 			 `json:"price" bson:"price"`
	Description   string   			 `json:"description,omitempty" bson:"description"`
	Quantity      int			 `json:"quantity" bson:"quantity"`
	Visibility    bool     			 `json:"-" bson:"visibility"`
	// Type defines whether it is a product or service which we can then filter to find either
	// a businesses products or services, allows any store to sell anything
}

type Price struct {
	Price       	float64 	`json:"price"`
	Discount    	float64 	`json:"discount,omitempty"`
	PreTaxTotal 	float64 	`json:"pre-tax-total,omitempty"`
	Tax				float64 	`json:"tax"`
	Total       	float64 	`json:"total"`
}

type Payments struct {
	ID 			    primitive.ObjectID  `bson:"_id,omitempty"`
	CustomerID		string 				`bson:"customer-id"`
	Status			string    			`bson:"status"`
	Gateway			string				`bson:"gateway"`
	Type 			string				`bson:"type"`
	Amount 			string				`bson:"amount"`
	Card 			Card				`bson:"card"`
	Token 			string				`bson:"token"`
}

type Card struct {
	Brand			string	`json:"brand"`
	PanLastFour		string	`json:"pan-last-four"`
	ExpirationMonth	string	`json:"exp-month"`
	ExpirationYear  string	`json:"exp-year"`
	CvvVerified     bool	`json:"cvv-verified"`

}

type OrderArgs struct {
	Sid		string	`json:"sid"`
	Uid    string	`json:"uid"`
	Status bool `json:"status"`
	Total  float64	`json:"total"`
	Items []Item	`json:"items"`
}

type OrderView struct {
	ID		string		`json:"id"`
	Store	Store		`json:"store"`
	Status  OrderStatus	`json:"status"`
	Created string	`json:"created"`
	Completed string `json:"completed"`
	Total    float64	`json:"total"`
	Items   []Item		`json:"items"`
	Driver  string		`json:"driver"`
}

type Order struct {
	ID 				primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`

	Sid			string					`json:"sid" bson:"sid"`
	StoreName   string					`json:"sname,omitempty" bson:"sname,omitempty"`
	StorePhone	string					`json:"sphone,omitempty" bson:"sphone,omitempty"`
	Source      GeoJSON					`json:"source,omitempty" bson:"source,omitempty"`

	Uid			string					`json:"uid" bson:"uid"`
	UserName    string					`json:"uname,omitempty" bson:"uname,omitempty"`
	UserPhone   string					`json:"uphone,omitempty" bson:"uphone,omitempty"`
	Destination GeoJSON					`json:"destination,omitempty" bson:"destination,omitempty"`

	PaymentStatus   bool				`json:"-" bson:"payment_status"`
	Status 			OrderStatus 		`json:"status" bson:"status"`
	Created 		time.Time			`json:"created" bson:"created"`
	Completed 		time.Time			`json:"completed,omitempty" bson:"completed"`
	Items       	[]Item			  	`json:"items" bson:"items"`
	Total			float64			`json:"total" bson:"total"`
	Driver		    string 			`json:"did" bson:"did"`
}

type Driver struct {
	ID 		  primitive.ObjectID	`json:"id,omitempty" bson:"_id,omitempty"`
	Name      string    `json:"name"`
	Email          string    `json:"email" validate:"required,email"`
	Phone 		   string    `json:"phone"`
	Location       GeoJSON	 `json:"location" bson:"location"`
	Password 	   string    `json:"password,omitempty" validate:"required"`
	//SSN 	  string 	`json:"SSN,omitempty"`
	//HashedSSN []byte    `json:"-"`
	HashedPassword  []byte 		       `json:"-"`
	//Insurance Insurance `json:"insurance,omitempty"`
	//License   License   `json:"license"`
	//Vehicle   Vehicle   `json:"vehicle,omitempty"`
	Created time.Time 	`json:"-"`
	Active 	bool		`json:"-" bson:"active"`
}

//Image data type for the image object
type Image struct {
	photo map[int]int
}

//Vehicle data type for the vehicle object
type Vehicle struct {
	//ID           primitive.ObjectID `bson:"_id,omitempty"`
	VehicleMake  string             `json:"vehicleMake"`
	VehicleModel string             `json:"vehicleModel"`
	VehicleYear  string             `json:"vehicleYear"`
	VehicleColor string             `json:"vehicleColor"`
	VinNumber    string             `json:"vinNumber"`
}

//Insurance data type for the insurance object
type Insurance struct {
	//ID                primitive.ObjectID `bson:"_id,omitempty"`
	//Insured           *User              `json:"insured"`
	InsuranceProvider string             `json:"insurance-provider,omitempty"`
	PolicyNumber      string             `json:"policy-number,omitempty"`
	ExpirationDate    string          `json:"expiration-date,omitempty"`
	//Vehicle           Vehicle            `json:"vehicle"`
}

//License data type for the license object
type License struct {
	Proof Image `json:"proof"`
}

