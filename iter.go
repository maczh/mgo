package mgo

import (
	"context"
	"errors"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

// Iter 模拟mgo.v2的Iter
type Iter struct {
	cursor *mongo.Cursor
	ctx    context.Context
	cancel context.CancelFunc
	err    error
	closed bool
	mu     sync.Mutex
}

// Next 模拟mgo.v2的Next方法
func (i *Iter) Next(result interface{}) bool {
	if i == nil {
		return false
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.err != nil {
		return false
	}
	if i.ctx == nil {
		i.ctx = context.Background()
	}

	if i.cursor == nil || i.closed || !i.cursor.Next(i.ctx) {
		if i.cursor == nil {
			i.err = errors.New("iterator is not initialized")
			return false
		}
		i.err = i.cursor.Err()
		_ = i.closeLocked()
		return false
	}

	i.err = i.cursor.Decode(result)
	return i.err == nil
}

// Err 模拟mgo.v2的Err方法
func (i *Iter) Err() error {
	if i == nil {
		return errors.New("iterator is not initialized")
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.err != nil {
		return i.err
	}

	if i.cursor == nil {
		return errors.New("iterator is not initialized")
	}
	return i.cursor.Err()
}

// Close 模拟mgo.v2的Close方法
func (i *Iter) Close() error {
	if i == nil {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.closeLocked()
}

func (i *Iter) closeLocked() error {
	if i.cursor == nil {
		return nil
	}
	if i.closed {
		return nil
	}
	i.closed = true

	if i.cancel != nil {
		i.cancel()
	}
	if i.ctx == nil {
		i.ctx = context.Background()
	}
	return i.cursor.Close(i.ctx)
}
