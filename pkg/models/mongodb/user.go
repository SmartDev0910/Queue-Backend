package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/CS-PCockrill/queue/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//UserFunctions is used to inject mongo driver client and context into the application
type UserFunctions struct {
	CLIENT *mongo.Client
}

//Insert method is used to insert new user into the User collection
func (u *UserFunctions) Insert(name, email, phone, password string, location models.GeoJSON) (primitive.ObjectID, error) {
	//Insert user to the database
	userCollection := u.CLIENT.Database("queue").Collection("users")
	var user models.User

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}
	currentTime := time.Now()

	user.Name = name
	user.Email = email
	user.Phone = phone
	user.HashedPassword = hashedPassword
	user.Created = currentTime
	user.Active = true
	user.Location = location

	//Insert the user into the database
	result, err := userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		//FIXME: This code should return an unsuccessful request not an empty string, it currently outputs record inserted...
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}

	//Check ID of the inserted document
	insertedID := result.InsertedID.(primitive.ObjectID)
	fmt.Println(insertedID)

	return insertedID, nil
}

func (u *UserFunctions) UpdateUser(id primitive.ObjectID, name, email, phone string) error {
	userCollection := u.CLIENT.Database("queue").Collection("users")

	var user models.User
	user.Name = name
	user.Email =  email
	user.Phone = phone

	filter := bson.M{"_id": id}
	_, err := userCollection.UpdateOne(
		context.TODO(),
		filter,
		bson.D{{Key: "$set", Value: bson.M{
			"name": user.Name,

			"email": user.Email,
			"phone": user.Phone},
		}})

	if err != nil {
		fmt.Println("UpdateOne() result ERROR:", err)
		return err
	}
	return nil
}

func (u *UserFunctions) UpdateAddress(id primitive.ObjectID, street, street2, city, state, country, zip string) error {
	userCollection := u.CLIENT.Database("queue").Collection("users")
	
	var address models.Address
	address.Street = street
	address.Suite = street2
	address.City = city
	address.State = state
	address.Country = country
	address.Zip = zip

	filter := bson.M{"_id": bson.M{"$eq": id}}

	_, err := userCollection.UpdateOne(
		context.TODO(),
		filter, 
		bson.D{{Key: "$set", Value: bson.M{"address": bson.M{
			"street": address.Street,
			"suite": address.Suite,
			"city": address.City,
			"state": address.State,
			"country": address.Country,
			"zip": address.Zip,
		}},
	}})

	if err != nil {
        fmt.Println("UpdateOne() result ERROR:", err)
        return err
    }
	return nil
}

//Authenticate method to confirm if a user exists in the database
func (u *UserFunctions) Authenticate(email, password string) (primitive.ObjectID, error) {
	//Authenticate user before login by retrieving the user id and hashed password from database
	//Hash the password entered and compare it to the one retrieved from database
	userCollection := u.CLIENT.Database("queue").Collection("users")

	var output models.User
	filter := bson.M{"email": email}

	ctx, _ := context.WithTimeout(context.Background(), 30 *time.Second)
	err := userCollection.FindOne(ctx, filter).Decode(&output)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return output.ID, models.ErrNoRecord
		}
		log.Fatal(err)
	}
	// Check whether the hashed password and plain-text password provided match

	err = bcrypt.CompareHashAndPassword(output.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return output.ID, models.ErrInvalidCredentials
		} else {
			return output.ID, err
		}
	}

	fmt.Println("Authenticated")
	return output.ID, nil
}

// GetUserFromDB takes a Request and return that user
// TODO: The implementation of isAuthenticated (in middleware) to prevent so many error checks on cookies and jwt
// Would also be able to authenticate client at time of request, not after
func (u *UserFunctions) GetUserFromDB(uid string) (models.User, error) {
	userCollection := u.CLIENT.Database("queue").Collection("users")
	id, _ := primitive.ObjectIDFromHex(uid)

	var newUser models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&newUser)
	if err != nil {
		return models.User{}, err
	}
	return newUser, nil
}

// GetUsers to return all users in the database
func (u *UserFunctions) GetUsers() []models.User {
	userCollection := u.CLIENT.Database("queue").Collection("users")
	var newUser []models.User

	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &newUser); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	return newUser
}

