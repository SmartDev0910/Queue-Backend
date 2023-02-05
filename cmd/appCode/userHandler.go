package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CS-PCockrill/queue/cmd/appCode/common"
	"github.com/CS-PCockrill/queue/pkg/models"
	"io"
	"log"
	"net/http"
)


func (app *appInjection) TestStatus(w http.ResponseWriter, r *http.Request) {
	uid, err := app.GetUIDFromToken(r)
	if err != nil {
		//app.JSON(w, http.StatusUnauthorized)
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, err := app.user.GetUserFromDB(uid)
	pread, pwrite := io.Pipe()

	go func() {
		fmt.Fprint(pwrite, id.Active)
		pwrite.Close()
	}()
	//cmd := exec.Command("active")

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(pread)
	fmt.Printf("Active: %s\n", buffer.String())

}

// RegisterUser registers/inserts a user to database
func (app *appInjection) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newUser models.User

	// Parse and decode the request body into a user struct
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	err := app.decodeJSONBody(w, r, &newUser)

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

	// Check if the password length is at least minimum required length
	if len(newUser.Password) < 6 {
		message := make(map[string]string)
		message["pass-valid"] = "Your password must be at least 6 characters."
		if !isEmailValid(newUser.Email) {
			message["email-valid"] = "The email you entered is invalid"
		}
		//app.JSON(w, message)
		app.JSON(w, Map{"success": false, "error": "invalid email"})
		return
	}
	// Check through email REGEX that the provided user email is valid
	if !isEmailValid(newUser.Email) {
		//app.JSON(w, Map{"email-valid": "The email you entered is invalid."})
		app.JSON(w, Map{"success": false, "error": "invalid email"})
		return
	}
	// If there are no form validation errors, create a new user from JSON payload
	// Insert the new user into the database
	uid, err := app.user.Insert(
		newUser.Name,
		newUser.Email,
		newUser.Phone,
		newUser.Password,
		newUser.Location)

	if err != nil {
		app.JSON(w, Map{"success": false, "user": nil})
		return
	}
	app.SetJWTCookie(w, uid)
	user, err := app.user.GetUserFromDB(uid.Hex())
	common.CheckError(err)
	app.JSON(w, Map{"success": true, "user": user})
	return
}

// LoginUser logs into a user
func (app *appInjection) LoginUser(w http.ResponseWriter, r *http.Request) {
	// we don't use parseForm because we are reading from a request body and not
	// a go template file
	w.Header().Set("Content-Type", "application/json")
	var user models.User
	err := app.decodeJSONBody(w, r, &user)

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
	// Check that user credentials are valid, if not send user a message,
	// then re-display the login page
	id, err := app.user.Authenticate(user.Email, user.Password)
	user, err = app.user.GetUserFromDB(id.Hex())
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	// Create JWT payload in order to create a token
	app.SetJWTCookie(w, id)
	app.JSON(w, Map{"success": true, "user": user})
	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte(`{ "message": "success"'}`))
}

// Logout logs out user/store
func (app *appInjection) Logout(w http.ResponseWriter, r *http.Request) {
	app.RemoveJWTCookie(w)
	app.JSON(w, Map{"message": "Logged out"})
}

func (app *appInjection) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	var newPayment models.Card
	// Decode input from request body into Card struct,
	payment := app.decodeJSONBody(w, r, &newPayment)
	if payment != nil {
		// if there's an error return bad request
		app.clientError(w, http.StatusBadRequest)
		return
	}

	uid, err := app.GetUIDFromToken(r)
	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, _ := app.user.GetUserFromDB(uid)
	app.JSON(w, Map{"user": id})
	// FIXME: Insert new card into the users' payment options
	//_, err := app.user.UpdatePayment(id, payment)
}


func (app *appInjection) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// FIXME: Input the existing fields in case user just updates 1 field doesn't set all the others to empty
	// FIXME: This may be just a front-end issue and fill existing field values with user's state values
	var user models.User
	//w.Header().Set("Content-Type", "application/json")
	err := app.decodeJSONBody(w, r, &user)
	// check and gracefully respond to errors
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

	uid, err := app.GetUIDFromToken(r)
	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, err := app.user.GetUserFromDB(uid)
	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		return
	}
	// TODO: Possibly use only 1 update method but calls different update based on type of user (User/Store/Driver)
	app.user.UpdateUser(id.ID, user.Name, user.Email, user.Phone)
}

// UpdateAddress We could make a function that accepts all the parameters of a user, and it fills the values if
// new ones are entered, and doesn't change the existing values
func (app *appInjection) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	var address models.Address
	//w.Header().Set("Content-Type", "application/json")
	err := app.decodeJSONBody(w, r, &address)
	// check and gracefully respond to errors
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

	uid, err := app.GetUIDFromToken(r)

	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	id, err := app.user.GetUserFromDB(uid)
	if err != nil {
		app.JSON(w, Map{"message": "unauthorized"})
		return
	}

	app.user.UpdateAddress(id.ID, address.Street, address.Suite, address.City, address.State, address.Country, address.Zip)
}

// GetUser gets the current store logged in
func (app *appInjection) GetUser(w http.ResponseWriter, r *http.Request) {
	uid, err := app.GetUIDFromToken(r)
	//fmt.Printf("UID : %s", uid)
	if err != nil {
		//app.JSON(w, http.StatusUnauthorized)
		app.clientError(w, http.StatusUnauthorized)

		return
	}

	id, err := app.user.GetUserFromDB(uid)
	app.JSON(w, id)
}


func (app *appInjection) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := app.user.GetUsers()
	json.NewEncoder(w).Encode(users)
}
