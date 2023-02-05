package mongodb

import (
	"context"
	"fmt"
	"github.com/CS-PCockrill/queue/cmd/appCode/common"
	"github.com/CS-PCockrill/queue/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
)

type OrderFunctions struct{
	CLIENT *mongo.Client
}

// Key FIXME: Temporary Solution to mapping locations to drivers efficiently with a (lat, long) tuple key, driver ID value
type Key struct {
	Lat, Long float64
}

func (o *OrderFunctions) CreateOrder(order *models.Order) (primitive.ObjectID, error) {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	//order.Status = models.Received


	result, err := orderCollection.InsertOne(context.TODO(), order)
	if err != nil {
		object, _ := primitive.ObjectIDFromHex("")
		return object, err
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	// Register a store after registering a user...
	//fmt.Println(insertedID)
	return insertedID, nil
}

func (o *OrderFunctions) GetOrders(uid string) ([]models.Order, error) {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	var orders []models.Order
	var cursor *mongo.Cursor
	var err error

	//intFilter, _ := strconv.ParseInt(filter, 0, 32)
	cursor, err = orderCollection.Find(context.TODO(), bson.M{"uid": bson.M{"$eq": uid}})

	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &orders); err != nil {
		log.Fatal(err)
	}
	return orders, nil
}

func (o *OrderFunctions)  UpdateOrderDriver(orderId primitive.ObjectID, driverId string) error {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	result, err := orderCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": bson.M{"$eq": orderId}},
		bson.D{
			{"$set", bson.D{{"did", driverId}}},
		})
	if err != nil {
		//log.Fatal(err)
		return err
	}
	fmt.Println(fmt.Sprintf("ORDER DEBUG: RESULT _ %v", result))
	return nil
}


func (o *OrderFunctions) UpdateOrderStatus(orderID primitive.ObjectID, status string) (models.Order, error ){
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	intStatus, _ := strconv.ParseInt(status, 0, 32)

	_, err := orderCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": bson.M{"$eq": orderID}},
		bson.D{
			{"$set", bson.D{{"status", intStatus}}},
		})

	if err != nil {
		return models.Order{}, err
	}

	order, err := o.GetOrder(orderID.Hex())
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (o *OrderFunctions) AssignDriver(orderID primitive.ObjectID) (models.Driver, error) {
	return models.Driver{}, nil
}

func (o *OrderFunctions) FilterOrderStatus(filter models.OrderStatus) []models.Order {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	var orders []models.Order
	var cursor *mongo.Cursor
	var err error

	//intFilter, _ := strconv.ParseInt(filter, 0, 32)
	cursor, err = orderCollection.Find(context.TODO(), bson.M{"status": bson.M{"$eq": filter}})

	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	if err = cursor.All(context.TODO(), &orders); err != nil {
		log.Fatal(err)
	}
	return orders
}

// GetUsers to return all users in the database
func (o *OrderFunctions) GetOrder(id string) (models.Order,error) {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	var order models.Order
	uid, _ := primitive.ObjectIDFromHex(id)

	err := orderCollection.FindOne(context.TODO(), bson.M{"_id": bson.M{"$eq": uid}}).Decode(&order)
	if err != nil {
		fmt.Println("ERROR HERE")
		return models.Order{}, err
	}
	return order, nil
}

type documentKey struct {
	ID primitive.ObjectID `bson:"_id"`
}

type changeID struct {
	Data string `bson:"_data"`
}

type namespace struct {
	Db   string `bson:"db"`
	Coll string `bson:"coll"`
}

// This is an example change event struct for inserts.
// It does not include all possible change event fields.
// You should consult the change event documentation for more info:
// https://docs.mongodb.com/manual/reference/change-events/
type changeEvent struct {
	ID            changeID            `bson:"_id"`
	OperationType string              `bson:"operationType"`
	ClusterTime   primitive.Timestamp `bson:"clusterTime"`
	FullDocument  models.Order               `bson:"fullDocument"`
	DocumentKey   documentKey         `bson:"documentKey"`
	Ns            namespace           `bson:"ns"`
}

func (o *OrderFunctions) AlertStore() models.Order {
	orderCollection := o.CLIENT.Database("queue").Collection("orders")
	orderStream, err := orderCollection.Watch(context.TODO(), mongo.Pipeline{})
	common.CheckError(err)

	defer orderStream.Close(context.TODO())
	for orderStream.Next(context.TODO()) {
		var data changeEvent
		var order models.Order
		//err := orderStream.Decode(&data)

		if err := orderStream.Decode(&data); err != nil {
			log.Fatal(err)
		}
		bsonBytes, _ := bson.Marshal(data.FullDocument)
		bson.Unmarshal(bsonBytes, &order)

		// TODO: Alert the Store with SID that order has been received
		fmt.Printf("%v\n", order.Sid)
		return order
	}
	return models.Order{}
}
