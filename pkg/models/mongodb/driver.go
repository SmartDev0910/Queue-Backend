package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/CS-PCockrill/queue/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type DriverFunctions struct{
	CLIENT *mongo.Client
}

func (d *DriverFunctions) CreateIndex() error {
	storeCollection := d.CLIENT.Database("queue").Collection("drivers")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	indexOpts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	pointIndexModel := mongo.IndexModel{
		Keys:    bsonx.MDoc{"location": bsonx.String("2dsphere")},
		Options: options.Index().SetBackground(true),
	}
	pointIndexes := storeCollection.Indexes()
	_, err := pointIndexes.CreateOne(
		ctx,
		pointIndexModel,
		indexOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DriverFunctions) RegisterDriver(driver *models.Driver) (primitive.ObjectID, error) {
	// Register a driver, verify background check, and verify state id & insurance
	DB := d.CLIENT.Database("queue")
	driverCollection := DB.Collection("drivers")
	//var newDriver models.Driver

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(driver.Password), 12)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}

	//hashedSSN, err := bcrypt.GenerateFromPassword([]byte(driver.SSN), 12)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}

	currentTime := time.Now()

	//driver.SSN = ""
	driver.Password = ""
	driver.HashedPassword = hashedPassword
	//driver.HashedSSN = hashedSSN
	driver.Created = currentTime
	driver.Active = true

	result, err := driverCollection.InsertOne(context.TODO(), driver)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	// Register a driver
	return insertedID, nil
	// Get the current user in session
	// driver.User = current session user
}


func (d *DriverFunctions) Authenticate(email, password string) (primitive.ObjectID, error) {
	//Authenticate user before login by retrieving the user id and hashed password from database
	//Hash the password entered and compare it to the one retrieved from database
	driverCollection := d.CLIENT.Database("queue").Collection("drivers")

	var output models.Driver
	filter := bson.M{"email": email}

	err := driverCollection.FindOne(context.TODO(), filter).Decode(&output)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return output.ID, errors.New(fmt.Sprintf("%v", err))
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

	fmt.Println("Authenticated Driver")
	return output.ID, nil
}

func (d *DriverFunctions) GetDriver(uid primitive.ObjectID) (models.Driver, error) {
	driverCollection := d.CLIENT.Database("queue").Collection("drivers")
	var output models.Driver
	//id, _ := primitive.ObjectIDFromHex(uid)
	filter := bson.M{"_id": bson.M{"$eq": uid}}
	err := driverCollection.FindOne(context.TODO(), filter).Decode(&output)
	if err != nil {
		return models.Driver{}, err
	}
	return output, nil

	//userCollection := u.CLIENT.Database("queue").Collection("users")
	//id, _ := primitive.ObjectIDFromHex(uid)
	//
	//var newUser models.User
	//err := userCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&newUser)
	//if err != nil {
	//	return models.User{}, err
	//}
	//return newUser, nil
}

//func (d *DriverFunctions) GetOnlineDrivers() []models.Driver {
//	driverCollection := d.CLIENT.Database("queue").Collection("drivers")
//	var drivers []models.Driver
//	//var cursor *mongo.Cursor
//
//	matchStage := bson.D{{"$match", bson.D{{"active", true}}}}
//	//excludeFields := bson.D{{"$project", bson.D{{"insurance", 0}, {"hashedssn", 0}}}}
//
//	cursor, err := driverCollection.Aggregate(context.Background(), mongo.Pipeline{
//		matchStage,
//		excludeFields,
//	})
//
//	defer cursor.Close(context.TODO())
//	if err = cursor.All(context.TODO(), &drivers); err != nil {
//		log.Fatal(err)
//	}
//	return drivers
//}
// GetPointsByDistance gets all the points that are within the
// maximum distance provided in meters.
func (d *DriverFunctions) GetPointsByDistance(location models.GeoJSON, distance int) ([]models.Driver, error) {
	coll := d.CLIENT.Database("queue").Collection("drivers")
	//var results []models.GeoJSON
	filter := bson.D{
		{"location",
			bson.D{
				{"$near", bson.D{
					{"$geometry", location},
					{"$maxDistance", distance},
				}},
			}},
	}

	var drivers []models.Driver
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &drivers); err != nil {
		log.Fatal(err)
	}
	return drivers, nil
}


func (d *DriverFunctions) GetDrivers() []models.Driver {
	userCollection := d.CLIENT.Database("queue").Collection("drivers")
	var drivers []models.Driver

	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &drivers); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	return drivers
}

func (d *DriverFunctions) Validate(email, password string) (int, error) {
	// Validate the drivers login credentials
	return 0, nil
}

func (d *DriverFunctions) getDriver(id int) *models.Driver {
	// Get a driver with parameter id
	return nil
}
