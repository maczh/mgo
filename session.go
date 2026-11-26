package mgo

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ownsClient    bool
}

var errSessionClosed = errors.New("session is closed")

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
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dbName, err := extractDBName(url)
	if err != nil {
		return nil, err
	}
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
		disconnectClient(client)
		return nil, err
	}
	sessionCtx, sessionCancel := context.WithCancel(context.Background())

	s := &Session{
		client:        client,
		ctx:           sessionCtx,
		cancel:        sessionCancel,
		mode:          Monotonic,
		safe:          &Safe{W: 1},
		poolLimit:     poolMax,
		socketTimeout: timeout,
		url:           url,
		database:      dbName,
		ownsClient:    true,
	}

	return s, nil
}

// Clone 模拟mgo.v2的Clone方法
func (s *Session) Clone() *Session {
	if s == nil {
		return nil
	}
	s.RLock()
	defer s.RUnlock()

	if s.isClosed {
		return nil
	}

	cloneCtx, cloneCancel := context.WithCancel(context.Background())
	newSession := &Session{
		client:        s.client,
		ctx:           cloneCtx,
		cancel:        cloneCancel,
		database:      s.database,
		mode:          s.mode,
		safe:          cloneSafe(s.safe),
		poolLimit:     s.poolLimit,
		socketTimeout: s.socketTimeout,
		ownsClient:    false,
	}

	return newSession
}

// Copy 模拟mgo.v2的Copy方法
func (s *Session) Copy() *Session {
	if s == nil {
		return nil
	}
	s.RLock()
	if s.isClosed {
		s.RUnlock()
		return nil
	}
	url, poolLimit, database, mode, safe, socketTimeout := s.url, s.poolLimit, s.database, s.mode, cloneSafe(s.safe), s.socketTimeout
	s.RUnlock()
	if socketTimeout <= 0 {
		socketTimeout = 10 * time.Second
	}

	opts := options.Client().ApplyURI(url).SetMaxPoolSize(uint64(poolLimit))

	connectCtx, connectCancel := context.WithTimeout(context.Background(), socketTimeout)
	defer connectCancel()
	newClient, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		if newClient != nil {
			disconnectClient(newClient)
		}
		return nil
	}

	copyCtx, copyCancel := context.WithCancel(context.Background())
	newSession := &Session{
		client:        newClient,
		ctx:           copyCtx,
		cancel:        copyCancel,
		database:      database,
		mode:          mode,
		safe:          safe,
		poolLimit:     poolLimit,
		socketTimeout: socketTimeout,
		ownsClient:    true,
	}

	return newSession
}

// DB 模拟mgo.v2的DB方法
func (s *Session) DB(name string) *Database {
	if s == nil {
		return nil
	}
	s.Lock()
	defer s.Unlock()
	if name != "" {
		s.database = name
	}

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
	if s == nil {
		return
	}
	s.Lock()
	defer s.Unlock()

	if safe == nil {
		s.safe = &Safe{W: 1}
	} else {
		s.safe = cloneSafe(safe)
	}
}

// Close 模拟mgo.v2的Close方法
func (s *Session) Close() {
	if s == nil {
		return
	}
	s.Lock()
	if s.isClosed {
		s.Unlock()
		return
	}
	s.isClosed = true
	client, cancel := s.client, s.cancel
	s.Unlock()
	if cancel != nil {
		cancel()
	}
	if client != nil && s.ownsClient {
		disconnectClient(client)
	}
}

// Ping 模拟mgo.v2的Ping方法
func (s *Session) Ping() error {
	if s == nil {
		return errSessionClosed
	}
	ctx, cancel, err := s.operationContext()
	if err != nil {
		return err
	}
	defer cancel()
	return s.client.Ping(ctx, nil)
}

func cloneSafe(safe *Safe) *Safe {
	if safe == nil {
		return &Safe{W: 1}
	}
	return &Safe{W: safe.W, WMode: safe.WMode, WTimeout: safe.WTimeout, FSync: safe.FSync, J: safe.J}
}

func disconnectClient(client *mongo.Client) {
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = client.Disconnect(ctx)
}

// SetSocketTimeout 模拟mgo.v2的SetSocketTimeout方法
func (s *Session) SetSocketTimeout(d time.Duration) {
	if s == nil {
		return
	}
	s.Lock()
	defer s.Unlock()

	s.socketTimeout = d
	// 需要设置到底层连接
}

// DatabaseNames 模拟mgo.v2的DatabaseNames
func (s *Session) DatabaseNames() (names []string, err error) {
	if s == nil {
		return nil, errSessionClosed
	}
	ctx, cancel, err := s.operationContext()
	if err != nil {
		return nil, err
	}
	defer cancel()
	return s.client.ListDatabaseNames(ctx, options.ListDatabases())
}

func (s *Session) operationContext() (context.Context, context.CancelFunc, error) {
	s.RLock()
	defer s.RUnlock()
	if s.isClosed || s.client == nil || s.ctx == nil {
		return nil, func() {}, errSessionClosed
	}
	timeout := s.socketTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	return ctx, cancel, nil
}
