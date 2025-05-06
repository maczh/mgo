package mgo

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

type user struct {
	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

func TestMgo(t *testing.T) {
	uri := "mongodb://test:711125@192.168.2.30:27017/test?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256"
	client, err := Dial(uri)
	if err != nil {
		t.Log(err)
		return
	}
	//var docs = []user{
	//	{Name: "user1", Age: 20},
	//	{Name: "user2", Age: 25},
	//	{Name: "user3", Age: 30},
	//}
	//err = client.DB("test").C("users").Insert(docs[2])
	//if err != nil {
	//	t.Log(err)
	//	return
	//}
	var result []user
	err = client.DB("test").C("users").Find(map[string]any{"age": bson.M{"$gte": 25}}).All(&result)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(result)
}
