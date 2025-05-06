package mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection 模拟mgo.v2的Collection
type Collection struct {
	session    *Session
	collection *mongo.Collection
	name       string
	database   *Database
}

// Insert 模拟mgo.v2的Insert方法
func (c *Collection) Insert(docs ...interface{}) error {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	if len(docs) == 0 {
		return nil
	}

	// 转换为bson.D切片
	bsonDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		bsonDoc, err := toBsonD(doc)
		if err != nil {
			return err
		}
		bsonDocs[i] = bsonDoc
	}

	_, err := c.collection.InsertMany(ctx, bsonDocs)
	return err
}

// Find 模拟mgo.v2的Find方法
func (c *Collection) Find(query interface{}) *Query {
	return &Query{
		session:    c.session,
		collection: c,
		query:      query,
		skip:       0,
		limit:      0,
		sort:       []string{},
	}
}

// Update 模拟mgo.v2的Update方法
func (c *Collection) Update(selector, update interface{}) error {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	selectorDoc, err := toBsonD(selector)
	if err != nil {
		return err
	}

	updateDoc, err := toBsonD(update)
	if err != nil {
		return err
	}

	_, err = c.collection.UpdateOne(ctx, selectorDoc, updateDoc)
	return err
}

type ChangeInfo struct {
	// Updated reports the number of existing documents modified.
	// Due to server limitations, this reports the same value as the Matched field when
	// talking to MongoDB <= 2.4 and on Upsert and Apply (findAndModify) operations.
	Updated    int
	Removed    int         // Number of documents removed
	Matched    int         // Number of documents matched but not necessarily changed
	UpsertedId interface{} // Upserted _id field, when not explicitly provided
}

// UpdateAll 模拟mgo.v2的UpdateAll方法
func (c *Collection) UpdateAll(selector interface{}, update interface{}) (*ChangeInfo, error) {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	selectorDoc, err := toBsonD(selector)
	if err != nil {
		return nil, err
	}

	updateDoc, err := toBsonD(update)
	if err != nil {
		return nil, err
	}

	result, err := c.collection.UpdateMany(ctx, selectorDoc, updateDoc)
	info := &ChangeInfo{
		Updated:    int(result.ModifiedCount),
		Removed:    0,
		Matched:    int(result.MatchedCount),
		UpsertedId: result.UpsertedID,
	}
	return info, err
}

// UpdateId 模拟mgo.v2的UpdateId方法
func (c *Collection) UpdateId(id, update interface{}) error {
	return c.Update(bson.M{"_id": id}, update)
}

// Upsert 模拟mgo.v2的Upsert方法
func (c *Collection) Upsert(selector, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	selectorDoc, err := toBsonD(selector)
	if err != nil {
		return nil, err
	}

	updateDoc, err := toBsonD(update)
	if err != nil {
		return nil, err
	}

	opts := options.Update().SetUpsert(true)
	return c.collection.UpdateOne(ctx, selectorDoc, updateDoc, opts)
}

// UpsertId 模拟mgo.v2的UpsertId方法
func (c *Collection) UpsertId(id, update interface{}) (*mongo.UpdateResult, error) {
	return c.Upsert(bson.M{"_id": id}, update)
}

// Remove 模拟mgo.v2的Remove方法
func (c *Collection) Remove(selector interface{}) error {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	selectorDoc, err := toBsonD(selector)
	if err != nil {
		return err
	}

	_, err = c.collection.DeleteOne(ctx, selectorDoc)
	return err
}

// RemoveId 模拟mgo.v2的RemoveId方法
func (c *Collection) RemoveId(id interface{}) error {
	return c.Remove(bson.M{"_id": id})
}

// RemoveAll 模拟mgo.v2的RemoveAll方法
func (c *Collection) RemoveAll(selector interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	selectorDoc, err := toBsonD(selector)
	if err != nil {
		return nil, err
	}

	return c.collection.DeleteMany(ctx, selectorDoc)
}

// Count 模拟mgo.v2的Count方法
func (c *Collection) Count() (int, error) {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	count, err := c.collection.CountDocuments(ctx, bson.D{})
	return int(count), err
}

// EnsureIndex 模拟mgo.v2的EnsureIndex方法
func (c *Collection) EnsureIndex(index Index) error {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	keys := bson.D{}
	for _, key := range index.Keys {
		direction := 1
		if key[0] == '-' {
			direction = -1
			key = key[1:]
		}
		keys = append(keys, bson.E{Key: key, Value: direction})
	}

	opts := options.Index()
	if index.Unique {
		opts = opts.SetUnique(true)
	}

	if index.Background {
		opts = opts.SetBackground(true)
	}

	if index.Sparse {
		opts = opts.SetSparse(true)
	}

	if index.ExpireAfterSeconds > 0 {
		opts = opts.SetExpireAfterSeconds(int32(index.ExpireAfterSeconds))
	}

	if index.Name != "" {
		opts = opts.SetName(index.Name)
	}

	model := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}

	_, err := c.collection.Indexes().CreateOne(ctx, model)
	return err
}

// DropCollection 模拟mgo.v2的DropCollection方法
func (c *Collection) DropCollection() error {
	ctx, cancel := context.WithTimeout(c.session.ctx, c.session.socketTimeout)
	defer cancel()

	return c.collection.Drop(ctx)
}

// toBsonD 将interface转换为bson.D
func toBsonD(data interface{}) (bson.D, error) {
	bsonBytes, err := bson.Marshal(data)
	if err != nil {
		return nil, err
	}

	var bsonDoc bson.D
	err = bson.Unmarshal(bsonBytes, &bsonDoc)
	if err != nil {
		return nil, err
	}

	return bsonDoc, nil
}

// Index 模拟mgo.v2的Index
type Index struct {
	Keys               []string
	Unique             bool
	Background         bool
	Sparse             bool
	ExpireAfterSeconds int
	Name               string
}
