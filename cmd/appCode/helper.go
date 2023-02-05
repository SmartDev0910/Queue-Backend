package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CS-PCockrill/queue/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/gddo/httputil/header"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	//"github.com/dgrijalva/jwt-go"
	"net/http"
	"runtime/debug"
)

type Map map[string]interface{}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func (app *appInjection) send(channel chan models.Order, value models.Order) {
	//var order models.Order
	//fmt.Printf("DEBUG: SEND: ", value)
	select {
	case channel <- value:
		//fmt.Printf("DEBUG: SEND: ", <-channel)
	default:
	}
}

func (app *appInjection) isAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	//_, err := app.user.GetUserFromDB(r)
	//if err != nil {
	//	fmt.Println("User is not authenticated: app isAuthenticated() false\n")
	//	return false
	//}

	return true
}

func (app *appInjection) GetUIDFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return "", err
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	//fmt.Printf("Err: %s\n", err)
	if err != nil || !token.Valid {
		return "", err
	}

	payload := token.Claims.(*jwt.StandardClaims)
	uid := payload.Subject
	return uid, nil
}

func (app *appInjection) SetJWTCookie(w http.ResponseWriter, id primitive.ObjectID) {
	payload := jwt.StandardClaims{
		Subject: primitive.ObjectID.Hex(id),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}

	// Create a token, and then pass it into a http.Cookie{}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(os.Getenv("SECRET_KEY")))
	cookie := &http.Cookie{
		Name: "jwt",
		Value: token,
		Path: "/",
		Expires: time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	}
	// Set session cookie and write success to response body
	http.SetCookie(w, cookie)
}

func (app *appInjection) RemoveJWTCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name: "jwt",
		Path: "/",
		Expires: time.Now().Add(-(10 * time.Minute)),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func (app *appInjection) decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	//if err != io.EOF {
	//	msg := "Request body must only contain a single JSON object"
	//	return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	//}

	return nil
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}
// JSON outputs a json message given parameter data to the response
func (app *appInjection) JSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

//func (app *appInjection) DistanceMatrix(pickup, currentLocation models.Address) error {
//	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/distancematrix/json?origins=%s&destinations=%s&units=imperial&key=AIzaSyDahbDRSHDuO-uaZ7jZ-9GXF9834xv_ZOk", pickup, currentLocation)
//	method := "GET"
//
//
//}

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *appInjection) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user.
func (app *appInjection) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *appInjection) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
