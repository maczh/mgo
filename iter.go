package mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

// Iter 模拟mgo.v2的Iter
type Iter struct {
	cursor *mongo.Cursor
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

// Next 模拟mgo.v2的Next方法
func (i *Iter) Next(result interface{}) bool {
	if i.err != nil {
		return false
	}

	if !i.cursor.Next(i.ctx) {
		i.err = i.cursor.Err()
		i.Close()
		return false
	}

	i.err = i.cursor.Decode(result)
	return i.err == nil
}

// Err 模拟mgo.v2的Err方法
func (i *Iter) Err() error {
	if i.err != nil {
		return i.err
	}

	return i.cursor.Err()
}

// Close 模拟mgo.v2的Close方法
func (i *Iter) Close() error {
	if i.cursor == nil {
		return nil
	}

	if i.cancel != nil {
		i.cancel()
	}

	return i.cursor.Close(i.ctx)
}
