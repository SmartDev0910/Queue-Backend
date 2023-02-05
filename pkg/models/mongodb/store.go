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

type StoreFunctions struct{
	CLIENT *mongo.Client
}

func (s *StoreFunctions) Insert(store *models.Store) (primitive.ObjectID, error) {
	//Insert store to the database
	storeCollection := s.CLIENT.Database("queue").Collection("stores")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(store.Password), 12)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}

	store.Password = ""
	store.HashedPassword = hashedPassword
	store.Inventory = []models.Item{}

	store.Location.Type = "Point"
	store.Location.Coordinates = []float64{38.91426220715202, -77.38484897262948}
	//store.Location.Coordinates = []float64{38.830938, -77.307314}
	//store.Location.Coordinates = []float64{38.825550972360645, -77.31454240136205}
	//myCoord := bson.D{{"type", "Point"}, {"coordinates", []float64{38.830938, -77.307314}}}

	result, err := storeCollection.InsertOne(context.TODO(), store)

	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	// Register a store after registering a user...
	return insertedID, nil
}

func NewPoint(long, lat float64) models.GeoJSON {
	return models.GeoJSON{
		Type:        "Point",
		Coordinates: []float64{long, lat},
	}
}

func (s *StoreFunctions) CreateIndex() error {
	storeCollection := s.CLIENT.Database("queue").Collection("stores")
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

// GetPointsByDistance gets all the points that are within the
// maximum distance provided in meters.
func (s *StoreFunctions) GetPointsByDistance(location models.GeoJSON, distance int) ([]models.Store, error) {
	coll := s.CLIENT.Database("queue").Collection("stores")
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

	var newStore []models.Store
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &newStore); err != nil {
		log.Fatal(err)
	}
	return newStore, nil
}


func (s *StoreFunctions) FilterLocation() []models.Store  {
	storeCollection := s.CLIENT.Database("queue").Collection("stores")
	myCoord := NewPoint(38.830938, -77.307314)
	filter := bson.D{
		{"location", bson.D{{
			"$near", bson.D{
				{"$geometry", myCoord},
				{"$maxDistance", 10000},
			},
		}},
		},
	}
	var stores []models.Store
	cursor, err := storeCollection.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &stores); err != nil {
		log.Fatal(err)
	}
	return stores
	//var stores []bson.D

	//if err = out.All(context.TODO(), &stores); err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(stores)
	//return stores
}

func (s *StoreFunctions) Authenticate(email, password string) (primitive.ObjectID, error) {
	//Authenticate user before login by retrieving the user id and hashed password from database
	//Hash the password entered and compare it to the one retrieved from database
	storeCollection := s.CLIENT.Database("queue").Collection("stores")

	var output models.Store
	filter := bson.M{"email": email}

	err := storeCollection.FindOne(context.TODO(), filter).Decode(&output)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return output.Id, errors.New("filter did not match any documents in the collection")
		}
		log.Fatal(err)
	}
	// Check whether the hashed password and plain-text password provided match
	err = bcrypt.CompareHashAndPassword(output.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return output.Id, models.ErrInvalidCredentials
		} else {
			return output.Id, err
		}
	}

	fmt.Println("Authenticated Store")
	return output.Id, nil
}

func (s *StoreFunctions) GetStoreFromDB(uid string) (models.Store, error) {
	storeCollection := s.CLIENT.Database("queue").Collection("stores")

	id, _ := primitive.ObjectIDFromHex(uid)

	var newStore models.Store
	err := storeCollection.FindOne(context.TODO(), bson.M{"_id": bson.M{"$eq": id}}).Decode(&newStore)
	if err != nil {
		return models.Store{}, err
	}
	return newStore, nil
}

//GetStores to get all stores
func (s *StoreFunctions) GetStores() []models.Store {
	userCollection := s.CLIENT.Database("queue").Collection("stores")
	var newStore []models.Store
	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &newStore); err != nil {
		log.Fatal(err)
	}
	return newStore
}

func (s *StoreFunctions) InsertOneProduct(id primitive.ObjectID, name, description string, price float64) (models.Item, error) {
	// Insert just one product if there is only 1
	storeCollection := s.CLIENT.Database("queue").Collection("stores")
	var newItem models.Item

	newItem.Id = primitive.NewObjectID()
	newItem.Name = name
	newItem.Price = price
	newItem.StoreId = primitive.ObjectID.Hex(id)
	newItem.Description = description
	newItem.Visibility = true

	filter := bson.M{"_id": bson.M{"$eq": id}}
	opts := options.Update().SetUpsert(true)

	_, err := storeCollection.UpdateOne(
		context.TODO(),
		filter,
		bson.M{"$push": bson.M{"inventory": newItem}}, opts,
	)

	if err != nil {
		fmt.Println("UpdateOne() result ERROR:", err)
		return models.Item{}, err
	}

	return newItem, nil
}

// UpdateOneItem will update item based on store id, and itemId.
// If store tries to update itemId that's not in their store, return err
func (s *StoreFunctions) UpdateOneItem(id string, itemId primitive.ObjectID, name, description string, price float64) (models.Item, error) {
	// Insert just one product if there is only 1
	itemCollection := s.CLIENT.Database("queue").Collection("stores")

	var newItem models.Item


	newItem.StoreId = id
	newItem.Name = name
	newItem.Price = price
	newItem.Description = description
	newItem.Visibility = false

	filter := bson.M{"store-id": bson.M{"$eq": id}, "_id": bson.M{"$eq": itemId}}
	opts := options.Update().SetUpsert(true)
	_, err := itemCollection.UpdateOne(
		context.TODO(),
		filter,
		bson.M{"$set": newItem}, opts,
		)

	if err != nil {
		return models.Item{}, err
	}
	return newItem, nil
}

func (s *StoreFunctions) GetStoreOrders(sid string) ([]models.Order, error) {
	orderCollection := s.CLIENT.Database("queue").Collection("orders")
	var orders []models.Order
	filter := bson.M{"sid": bson.M{"$eq": sid}}
	cursor, err := orderCollection.Find(context.TODO(), filter)
	if err != nil {
		return []models.Order{}, err
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &orders); err != nil {
		return []models.Order{}, err
	}
	return orders, nil
}

func (s *StoreFunctions) InsertManyProducts(products ...*models.Item) (int, error) {
	// Insert many products if there is an array of products
	return 0, nil
}

