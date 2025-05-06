package mgo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Session 模拟mgo.v2的Session
type Session struct {
	client   *mongo.Client
	ctx      context.Context
	cancel   context.CancelFunc
	database string
	mode     Mode
	safe     *Safe
	sync.RWMutex
	poolLimit     int
	socketTimeout time.Duration
	isClosed      bool
	url           string
}

// Mode 模拟mgo.v2的查询模式
type Mode int

const (
	// 各种Mode常量定义
	Monotonic Mode = iota
	Eventual
	Strong
)

// Safe 模拟mgo.v2的安全模式
type Safe struct {
	W        int
	WMode    string
	WTimeout int
	FSync    bool
	J        bool
}

func extractDBName(connectionString string) (string, error) {
	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return "", err
	}
	// 去除路径开头的斜杠
	path := parsedURL.Path[1:]
	return path, nil
}

// Dial 模拟mgo.v2的Dial函数
func Dial(url string) (*Session, error) {
	return DialWithTimeout(url, 10*time.Second, 0)
}

// DialWithTimeout 模拟mgo.v2的DialWithTimeout函数
func DialWithTimeout(url string, timeout time.Duration, poolMax int) (*Session, error) {
	if url == "" || !strings.HasPrefix(url, "mongodb://") {
		return nil, errors.New("mongodb empty url or invalid url")
	}
	dbName, err := extractDBName(url)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if poolMax <= 0 {
		poolMax = 10
	}
	opts := options.Client().ApplyURI(url).SetMaxPoolSize(uint64(poolMax))

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	s := &Session{
		client:        client,
		ctx:           context.Background(),
		cancel:        cancel,
		mode:          Monotonic,
		safe:          &Safe{W: 1},
		poolLimit:     4096,
		socketTimeout: timeout,
		url:           url,
		database:      dbName,
	}

	return s, nil
}

// Clone 模拟mgo.v2的Clone方法
func (s *Session) Clone() *Session {
	s.RLock()
	defer s.RUnlock()

	if s.isClosed {
		return nil
	}

	newSession := &Session{
		client:   s.client,
		ctx:      context.Background(),
		cancel:   s.cancel,
		database: s.database,
		mode:     s.mode,
		safe: &Safe{
			W:        s.safe.W,
			WMode:    s.safe.WMode,
			WTimeout: s.safe.WTimeout,
			FSync:    s.safe.FSync,
			J:        s.safe.J,
		},
		poolLimit:     s.poolLimit,
		socketTimeout: s.socketTimeout,
	}

	return newSession
}

// Copy 模拟mgo.v2的Copy方法
func (s *Session) Copy() *Session {
	s.RLock()
	defer s.RUnlock()

	if s.isClosed {
		return nil
	}

	opts := options.Client().ApplyURI(s.url).SetMaxPoolSize(uint64(s.poolLimit))

	newClient, err := mongo.Connect(s.ctx, opts)
	if err != nil {
		return nil
	}

	newSession := &Session{
		client:   newClient,
		ctx:      context.Background(),
		cancel:   s.cancel,
		database: s.database,
		mode:     s.mode,
		safe: &Safe{
			W:        s.safe.W,
			WMode:    s.safe.WMode,
			WTimeout: s.safe.WTimeout,
			FSync:    s.safe.FSync,
			J:        s.safe.J,
		},
		poolLimit:     s.poolLimit,
		socketTimeout: s.socketTimeout,
	}

	return newSession
}

// DB 模拟mgo.v2的DB方法
func (s *Session) DB(name string) *Database {
	if name != "" {
		s.database = name
	}
	s.RLock()
	defer s.RUnlock()

	if s.isClosed {
		return nil
	}

	db := s.client.Database(s.database)
	return &Database{
		session: s,
		db:      db,
		name:    s.database,
	}
}

// SetSafe 模拟mgo.v2的SetSafe方法
func (s *Session) SetSafe(safe *Safe) {
	s.Lock()
	defer s.Unlock()

	if safe == nil {
		s.safe = &Safe{W: 1}
	} else {
		s.safe = safe
	}
}

// Close 模拟mgo.v2的Close方法
func (s *Session) Close() {
	s.Lock()
	defer s.Unlock()

	if s.isClosed {
		return
	}

	s.client.Disconnect(s.ctx)
	s.isClosed = true
}

// Ping 模拟mgo.v2的Ping方法
func (s *Session) Ping() error {
	s.RLock()
	defer s.RUnlock()

	if s.isClosed {
		return errors.New("session is closed")
	}

	return s.client.Ping(s.ctx, nil)
}

// SetSocketTimeout 模拟mgo.v2的SetSocketTimeout方法
func (s *Session) SetSocketTimeout(d time.Duration) {
	s.Lock()
	defer s.Unlock()

	s.socketTimeout = d
	// 需要设置到底层连接
}

// DatabaseNames 模拟mgo.v2的DatabaseNames
func (s *Session) DatabaseNames() (names []string, err error) {
	return s.client.ListDatabaseNames(s.ctx, options.ListDatabases())
}
