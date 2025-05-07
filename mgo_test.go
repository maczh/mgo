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
	var docs = []any{
		user{Name: "user7", Age: 18},
		user{Name: "user8", Age: 23},
		user{Name: "user9", Age: 26},
	}
	err = client.DB("test").C("users").Insert(docs...)
	if err != nil {
		t.Log(err)
		return
	}
	db := client.DB("").C("users")
	var result []user
	query := db.Find(map[string]any{"age": bson.M{"$gte": 25}})
	err = query.Skip(1).Limit(2).Sort("age").All(&result)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(result)
	count, err := query.Limit(0).Skip(0).Count()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("总记录数：%d", count)
}
