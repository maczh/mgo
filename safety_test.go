package mgo

import (
	"context"
	"sync"
	"testing"
)

func TestOperationContextRejectsClosedSession(t *testing.T) {
	tests := []struct {
		name   string
		closed bool
		client bool
	}{
		{name: "closed", closed: true, client: true},
		{name: "missing client", client: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{isClosed: tt.closed, ctx: context.Background(), socketTimeout: 1}
			if !tt.client {
				s.client = nil
			}
			_, cancel, err := s.operationContext()
			cancel()
			if err != errSessionClosed {
				t.Fatalf("operationContext() error = %v, want %v", err, errSessionClosed)
			}
		})
	}
}

func TestEmptyIteratorDoesNotPanic(t *testing.T) {
	it := &Iter{}
	if it.Next(nil) {
		t.Fatal("Next() = true for an uninitialized iterator")
	}
	if it.Err() == nil {
		t.Fatal("Err() = nil for an uninitialized iterator")
	}
	if err := it.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestClosedDatabaseChainReturnsError(t *testing.T) {
	c := (&Database{}).C("users")
	if c == nil {
		t.Fatal("C() returned nil for an invalid database")
	}
	if err := c.Insert(map[string]string{"name": "test"}); err != errSessionClosed {
		t.Fatalf("Insert() error = %v, want %v", err, errSessionClosed)
	}
	q := c.Find(nil)
	if _, err := q.Count(); err != errSessionClosed {
		t.Fatalf("Count() error = %v, want %v", err, errSessionClosed)
	}
}

func TestNilReceiversAreSafe(t *testing.T) {
	var session *Session
	names, err := session.DatabaseNames()
	if session.Ping() != errSessionClosed || err != errSessionClosed || names != nil {
		t.Fatal("nil Session methods did not return safe results")
	}
	var iterator *Iter
	if iterator.Next(nil) || iterator.Err() == nil || iterator.Close() != nil {
		t.Fatal("nil Iter methods did not return safe results")
	}
}

func TestQueryConfigurationIsConcurrentSafe(t *testing.T) {
	q := (&Collection{}).Find(nil)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q.Skip(i).Limit(i + 1).Sort("name").Select(map[string]int{"name": 1})
		}(i)
	}
	wg.Wait()
}
