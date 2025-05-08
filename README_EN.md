# SDK Package for github.com/maczh/mgo

[Chinese Version](README.md "Chinese")

An imitation of the gopkg.in/mgo.v2 driver package, based on the official MongoDB Go language driver package.

## I. Overview

Since the gopkg.in/mgo.v2 package has not been updated for a long time and only supports MongoDB versions up to 4.4.x, but mgo.v2 is very well - encapsulated and easier to use than the official package. Therefore, an imitation mgo package is encapsulated using the official package, which can support the latest version of MongoDB by imitating the main commonly used functions of mgo.v2.

Usage example:

```go
package main

import (
    "go.mongodb.org/mongo-driver/bson"
    "github.com/maczh/mgo"
    "fmt"
)

type user struct {
    Name string `bson:"name"`
    Age  int    `bson:"age"`
}

func main() {
    uri := "mongodb://user:password@localhost:27017/testdb?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256"
    // Create a MongoDB connection
    client, err := mgo.Dial(uri)
    if err != nil {
        t.Log(err)
        return
    }
    var docs = []any{
        user{Name: "user7", Age: 18},
        user{Name: "user8", Age: 23},
        user{Name: "user9", Age: 26},
    }
    // Batch insert data
    err = client.DB("testdb").C("users").Insert(docs...)
    if err != nil {
        fmt.Println(err)
        return
    }
    // When the URL string contains the database name, DB("") directly uses the database in the URL string
    db := client.DB("").C("users")
    var result []user
    // Query data with pagination
    query := db.Find(map[string]any{"age": bson.M{"$gte": 25}})
    err = query.Skip(1).Limit(2).Sort("age").All(&result)
    if err != nil {
        fmt.Println(err)
        return
    }
    t.Log(result)
    // Get the total number of records according to the previous query statement
    count, err := query.CountAll()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("Total number of records: %d\n", count)
}
```

## II. Detailed Introduction of the SDK

### 2.1 Public Functions

#### func Dial() function

```go
func Dial(url string) (*Session, error)
```

Establish a database connection based on the MongoDB URL connection string. The format of the URL string is as follows:

```
mongodb://[user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
```

Explanation of options:

```
  connect=direct

       Disables the automatic replica set server discovery logic, and
       forces the use of servers provided only (even if secondaries).
       Note that to talk to a secondary the consistency requirements
       must be relaxed to Monotonic or Eventual via SetMode.

   connect=replicaSet

       Discover replica sets automatically. Default connection behavior.

   replicaSet=<setname>

       If specified will prevent the obtained session from communicating
       with any server which is not part of a replica set with the given name.
       The default is to communicate with any server specified or discovered
       via the servers contacted.

   authSource=<db>

       Informs the database used to establish credentials and privileges
       with a MongoDB server. Defaults to the database name provided via
       the URL path, and "admin" if that's unset.

   authMechanism=<mechanism>

      Defines the protocol for credential negotiation. Defaults to "MONGODB-CR",
      which is the default username/password challenge-response mechanism.

   gssapiServiceName=<name>

      Defines the service name to use when authenticating with the GSSAPI
      mechanism. Defaults to "mongodb".

   maxPoolSize=<limit>

      Defines the per-server socket pool limit. Defaults to 10.
```

1. Single-node database connection string: *mongodb://user:password@localhost:27017/testdb?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256*

2. Cluster database connection string: *mongodb://user:password@server1:27017,server2:27017/testdb?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256&replicaSet=repSet1*

#### func DialWithTimeout() function

```go
func DialWithTimeout(url string, timeout time.Duration, poolMax int) (*Session, error)
```

Connect to the MongoDB database with parameters. Parameters:

| Parameter Name | Parameter Description              |
| -------------- | ---------------------------------- |
| url            | MongoDB connection string          |
| timeout        | Connection timeout, minimum 10 seconds |
| poolMax        | Maximum value of the connection pool, default minimum is 10 |

#### func ToAnySlice() function

```go
func ToAnySlice(docs any) []any
```

Convert an array of structure objects to a []any function.

### 2.2 Session Class

Database connection session class.

#### func (s *Session) Clone() function

```go
func (s *Session) Clone() *Session
```

Clone a MongoDB connection.

#### func (s *Session) Copy() function

```go
func (s *Session) Copy() *Session
```

Copy a MongoDB connection. The copied connection is a new connection.

#### func (s *Session) DB() function

```go
func (s *Session) DB(name string) *Database
```

Select the database with the specified name. Returns a *Database object, equivalent to the use command.

If the name parameter is an empty string, select the database in the connection string.

#### func (s *Session) Close() function

```go
func (s *Session) Close()
```

Close the connection.

#### func (s *Session) Ping() function

```go
func (s *Session) Ping() error
```

Detect whether the database connection is normal.

#### func (s *Session) DatabaseNames() function

```go
func (s *Session) DatabaseNames() (names []string, err error)
```

Get the names of all databases in MongoDB, equivalent to the show dbs command.

### 2.3 Database Class

Database class, used for operations on a database with a specified name.

#### func (d *Database) C(name string) function

```go
func (d *Database) C(name string) *Collection
```

Specify the name of the Collection table and obtain the table object *Collection. If the specified table has not been created, it will be automatically created when the first record is inserted.

#### func (d *Database) Session() function

```go
func (d *Database) Session() *Session
```

Get the connection session object corresponding to the current Database object.

#### func (db *Database) CollectionNames() function

```go
func (db *Database) CollectionNames() (names []string, err error)
```

Get the names of all collection tables in the current database.

#### func (d *Database) AddUser() function

```go
func (d *Database) AddUser(username, password string, readOnly bool) error
```

Add a user to the database. The current connected user must have management permissions for the current database. The newly added user has only two permissions for this database: read-only and read-write.

| Parameter | Parameter Description           |
| --------- | ------------------------------- |
| username  | Username                        |
| password  | Login password                  |
| readOnly  | Whether it is read-only, true - read-only, false - read-write |

#### func (d *Database) RemoveUser() function

```go
func (d *Database) RemoveUser(username string) error
```

Delete a user from the database.

#### func (d *Database) DropDatabase() function

```go
func (d *Database) DropDatabase() error
```

Delete the current database.

### 2.4 Collection Class

Collection table object, used for all operations on the current table.

#### func (c *Collection) Database() function

```go
func (c *Collection) Database() *Database
```

Get the database object corresponding to the current table.

#### func (c *Collection) DropCollection() function

```go
func (c *Collection) DropCollection() error
```

Delete the current table.

#### func (c *Collection) Insert() function

```go
func (c *Collection) Insert(docs ...any) error
```

Insert data. You can insert one record or batch insert multiple records.

#### func (c *Collection) Update() function

```go
func (c *Collection) Update(selector, update interface{}) error
```

Update one record.

| Parameter | Parameter Description |
| --------- | --------------------- |
| selector  | Query condition       |
| update    | Fields to be updated  |

#### func (c *Collection) Find() function

```go
func (c *Collection) Find(query interface{}) *Query
```

The query statement returns a *Query object.

#### func (c *Collection) UpdateAll() function

```go
func (c *Collection) UpdateAll(selector interface{}, update interface{}) (*ChangeInfo, error)
```

Batch update records.

#### func (c *Collection) UpdateId() function

```go
func (c *Collection) UpdateId(id, update interface{}) error
```

Update a record by ID.

#### func (c *Collection) Upsert() function

```go
func (c *Collection) Upsert(selector, update interface{}) (*mongo.UpdateResult, error)
```

Insert or update function.

#### func (c *Collection) UpsertId() function

```go
func (c *Collection) UpsertId(id, update interface{}) (*mongo.UpdateResult, error)
```

Insert or update by ID.

#### func (c *Collection) Remove() function

```go
func (c *Collection) Remove(selector interface{}) error
```

Delete the first record that meets the conditions.

#### func (c *Collection) RemoveId() function

```go
func (c *Collection) RemoveId(id interface{}) error
```

Delete a record by ID.

#### func (c *Collection) RemoveAll() function

```go
func (c *Collection) RemoveAll(selector interface{}) (*mongo.DeleteResult, error)
```

Delete all records that meet the conditions.

#### func (c *Collection) Count() function

```go
func (c *Collection) Count() (int, error)
```

Count the total number of records in the table.

#### func (c *Collection) EnsureIndex() function

```go
func (c *Collection) EnsureIndex(index Index) error
```

Create an index.

### 2.5 Query Class

Dedicated to query processing with chained operations.

#### func (q *Query) Collection() function

```go
func (q *Query) Collection() *Collection
```

Return the parent table object.

#### func (q *Query) Select() function

```go
func (q *Query) Select(selector interface{}) *Query
```

Select the fields to be returned.

#### func (q *Query) Skip() function

```go
func (q *Query) Skip(n int) *Query
```

Skip a certain number of records, used for pagination.

#### func (q *Query) Limit() function

```go
func (q *Query) Limit(n int) *Query
```

Return the maximum number of records, used for pagination.

#### func (q *Query) Sort() function

```go
func (q *Query) Sort(fields ...string) *Query
```

Sort by the specified fields. Sorting takes precedence over Skip and Limit.

#### func (q *Query) One() function

```go
func (q *Query) One(result interface{}) error
```

Query and return the first record.

#### func (q *Query) All() function

```go
func (q *Query) All(result interface{}) error
```

Return all records that meet the conditions, including pagination control conditions.

#### func (q *Query) Count() function

```go
func (q *Query) Count() (int, error)
```

Count the number of records that meet the conditions, including Skip and Limit controls.

#### func (q *Query) CountAll() function

```go
func (q *Query) CountAll() (int, error)
```

Count the total number of records that meet the conditions. Skip and Limit do not take effect.