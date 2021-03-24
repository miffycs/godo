package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"server/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection object/instance
var collection *mongo.Collection

// create connection with MongoDB
func init() {
	loadTheEnv()
	createDBInstance()
}

// load .env file
func loadTheEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

// create DB instance
func createDBInstance() {
	// DB connection string
	connectionString := os.Getenv("DB_URI")
	// Database Name
	dbName := os.Getenv("DB_NAME")
	// Collection name
	collectionName := os.Getenv("DB_COLLECTION_NAME")

	// set client options
	clientOptions := options.Client().ApplyURI(connectionString)

	// connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// check connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = client.Database(dbName).Collection(collectionName)

	fmt.Println("Collection instance created!")
}

// GetAllTask: get all the task route
func GetAllTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	payload := getAllTask()
	json.NewEncoder(w).Encode(payload)
}

// CreateTask: create task route
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	var task models.ToDoList
	_ = json.NewDecoder(r.Body).Decode(&task)
	// fmt.Println(task, r.Body)
	insertOneTask(task)
	json.NewEncoder(w).Encode(task)
}

// TaskComplete: update task route
func TaskComplete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	taskComplete(params["id"])
	json.NewEncoder(w).Encode(params["id"])
}

// UndoTask: undo the complete task route
func UndoTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	undoTask(params["id"])
	json.NewEncoder(w).Encode(params["id"])
}

// DeleteTask: delete one task route
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	params := mux.Vars(r)
	deleteOneTask(params["id"])
	json.NewEncoder(w).Encode(params["id"])
	//json.NewEncoder(w).Encode("Task not found")
}

// DeleteAllTask: delete all tasks route
func DeleteAllTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	count := deleteAllTask()
	json.NewEncoder(w).Encode(count)
	json.NewEncoder(w).Encode("Task not found")
}

// getAllTask: get all task from DB and return
func getAllTask() []primitive.M {
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	var results []primitive.M
	for cur.Next(context.Background()) {
		var result bson.M
		e := cur.Decode(&result)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Println("cur..>", cur, "result", reflect.TypeOf(result), reflect.TypeOf(result["id"]))
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	cur.Close(context.Background())
	return results
}

// insertOneTask:
func insertOneTask(task models.ToDoList) {
	insertResult, err := collection.InsertOne(context.Background(), task)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single record ", insertResult.InsertedID)
}

// TODO: taskComplete and undoTask is the same other than T/F

// taskComplete: task complete method, update task's status to true
func taskComplete(task string) {
	fmt.Println(task)
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": true}}
	result, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("modified count: ", result.ModifiedCount)
}

// undoTask: task undo method, update task's status to false
func undoTask(task string) {
	fmt.Println(task)
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": false}}
	result, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("modified count: ", result.ModifiedCount)
}

// deleteOneTask: delete one task from the DB, delete by ID
func deleteOneTask(task string) {
	fmt.Println(task)
	id, _ := primitive.ObjectIDFromHex(task)
	filter := bson.M{"_id": id}
	d, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted document: ", d.DeletedCount)
}

// deleteAllTask: delete all the tasks from DB
func deleteAllTask() int64 {
	d, err := collection.DeleteMany(context.Background(), bson.D{{}}, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deleted document", d.DeletedCount)
	return d.DeletedCount
}
