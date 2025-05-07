package mgo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Query 模拟mgo.v2的Query
type Query struct {
	session      *Session
	collection   *Collection
	query        interface{}
	skip         int64
	limit        int64
	sort         []string
	selectFields interface{}
}

// Collection 获取Query所属的Collection
func (q *Query) Collection() *Collection {
	return q.collection
}

// Skip 模拟mgo.v2的Skip方法
func (q *Query) Skip(n int) *Query {
	q.skip = int64(n)
	return q
}

// Limit 模拟mgo.v2的Limit方法
func (q *Query) Limit(n int) *Query {
	q.limit = int64(n)
	return q
}

// Sort 模拟mgo.v2的Sort方法
func (q *Query) Sort(fields ...string) *Query {
	q.sort = fields
	return q
}

// Select 模拟mgo.v2的Select方法
func (q *Query) Select(selector interface{}) *Query {
	q.selectFields = selector
	return q
}

// One 模拟mgo.v2的One方法
func (q *Query) One(result interface{}) error {
	ctx, cancel := context.WithTimeout(q.session.ctx, q.session.socketTimeout)
	defer cancel()

	queryDoc, err := toBsonD(q.query)
	if err != nil {
		return err
	}

	opts := options.FindOne()
	opts.SetSkip(q.skip)

	if len(q.sort) > 0 {
		sortDoc := bson.D{}
		for _, field := range q.sort {
			direction := 1
			if field[0] == '-' {
				direction = -1
				field = field[1:]
			}
			sortDoc = append(sortDoc, bson.E{Key: field, Value: direction})
		}
		opts.SetSort(sortDoc)
	}

	if q.selectFields != nil {
		projectionDoc, err := toBsonD(q.selectFields)
		if err != nil {
			return err
		}
		opts.SetProjection(projectionDoc)
	}

	singleResult := q.collection.collection.FindOne(ctx, queryDoc, opts)
	if singleResult.Err() != nil {
		if singleResult.Err() == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return singleResult.Err()
	}

	return singleResult.Decode(result)
}

// All 模拟mgo.v2的All方法
func (q *Query) All(result interface{}) error {
	ctx, cancel := context.WithTimeout(q.session.ctx, q.session.socketTimeout)
	defer cancel()

	queryDoc, err := toBsonD(q.query)
	if err != nil {
		return err
	}

	opts := options.Find()
	opts.SetSkip(q.skip)

	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}

	if len(q.sort) > 0 {
		sortDoc := bson.D{}
		for _, field := range q.sort {
			direction := 1
			if field[0] == '-' {
				direction = -1
				field = field[1:]
			}
			sortDoc = append(sortDoc, bson.E{Key: field, Value: direction})
		}
		opts.SetSort(sortDoc)
	}

	if q.selectFields != nil {
		projectionDoc, err := toBsonD(q.selectFields)
		if err != nil {
			return err
		}
		opts.SetProjection(projectionDoc)
	}

	cursor, err := q.collection.collection.Find(ctx, queryDoc, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	err = cursor.All(q.session.ctx, result)
	return err
}

// Count 模拟mgo.v2的Count方法
func (q *Query) Count() (int, error) {
	ctx, cancel := context.WithTimeout(q.session.ctx, q.session.socketTimeout)
	defer cancel()

	queryDoc, err := toBsonD(q.query)
	if err != nil {
		return 0, err
	}

	opts := options.Count()
	opts.SetSkip(q.skip)

	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}

	count, err := q.collection.collection.CountDocuments(ctx, queryDoc, opts)
	return int(count), err
}

// CountAll 符合所有条件的记录数
func (q *Query) CountAll() (int, error) {
	return q.Skip(0).Limit(0).Count()
}

// Iter 模拟mgo.v2的Iter方法
func (q *Query) Iter() *Iter {
	ctx, cancel := context.WithTimeout(q.session.ctx, q.session.socketTimeout)

	queryDoc, err := toBsonD(q.query)
	if err != nil {
		return &Iter{
			err: err,
		}
	}

	opts := options.Find()
	opts.SetSkip(q.skip)

	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}

	if len(q.sort) > 0 {
		sortDoc := bson.D{}
		for _, field := range q.sort {
			direction := 1
			if field[0] == '-' {
				direction = -1
				field = field[1:]
			}
			sortDoc = append(sortDoc, bson.E{Key: field, Value: direction})
		}
		opts.SetSort(sortDoc)
	}

	if q.selectFields != nil {
		projectionDoc, err := toBsonD(q.selectFields)
		if err != nil {
			return &Iter{
				err: err,
			}
		}
		opts.SetProjection(projectionDoc)
	}

	cursor, err := q.collection.collection.Find(ctx, queryDoc, opts)
	if err != nil {
		return &Iter{
			err: err,
		}
	}

	return &Iter{
		cursor: cursor,
		ctx:    ctx,
		cancel: cancel,
	}
}

// ErrNotFound 模拟mgo.v2的ErrNotFound
var ErrNotFound = errors.New("not found")
