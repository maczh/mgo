package mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Database 模拟mgo.v2的Database
type Database struct {
	session *Session
	db      *mongo.Database
	name    string
}

// C 模拟mgo.v2的C方法
func (d *Database) C(name string) *Collection {
	return &Collection{
		session:    d.session,
		collection: d.db.Collection(name),
		name:       name,
		database:   d,
	}
}

// Session 获取Database所属的Session
func (d *Database) Session() *Session {
	return d.session
}

// DropDatabase 模拟mgo.v2的DropDatabase方法
func (d *Database) DropDatabase() error {
	ctx, cancel := context.WithTimeout(d.session.ctx, d.session.socketTimeout)
	defer cancel()

	return d.db.Drop(ctx)
}

// Run 模拟mgo.v2的Run方法
func (d *Database) Run(cmd interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(d.session.ctx, d.session.socketTimeout)
	defer cancel()

	cmdDoc := bson.M{}
	bsonBytes, err := bson.Marshal(cmd)
	if err != nil {
		return err
	}

	err = bson.Unmarshal(bsonBytes, &cmdDoc)
	if err != nil {
		return err
	}

	res := d.db.RunCommand(ctx, cmdDoc)
	if res.Err() != nil {
		return res.Err()
	}

	if result != nil {
		return res.Decode(result)
	}

	return nil
}

// AddUser 模拟mgo.v2的AddUser方法
func (d *Database) AddUser(username, password string, readOnly bool) error {
	cmd := bson.M{
		"createUser": username,
		"pwd":        password,
		"roles":      []bson.M{},
	}

	if readOnly {
		cmd["roles"] = []bson.M{{"role": "read", "db": d.name}}
	} else {
		cmd["roles"] = []bson.M{{"role": "readWrite", "db": d.name}}
	}

	return d.Run(cmd, nil)
}

// RemoveUser 模拟mgo.v2的RemoveUser方法
func (d *Database) RemoveUser(username string) error {
	cmd := bson.M{
		"dropUser": username,
	}

	return d.Run(cmd, nil)
}

// CollectionNames 模拟mgo.v2的CollectionNames方法
func (db *Database) CollectionNames() (names []string, err error) {
	return db.db.ListCollectionNames(context.Background(), bson.D{})
}
