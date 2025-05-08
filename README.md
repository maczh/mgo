# github.com/maczh/mgo SDK包

[English Version](README_EN.md)

仿gopkg.in/mgo.v2驱动包，基于mongodb官方go语言驱动包。

## 一、概览

由于gopkg.in/mgo.v2包长期未更新，对mongodb的版本仅支持到4.4.x版本，但是mgo.v2封装得非常好用，比官方包要易用，因此仿照mgo.v2的主要常用功能利用官方包封装出一个仿mgo包，可以支持最新版本的mongodb。

使用范例：

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
	//创建mongodb连接
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
    //批量插入数据
	err = client.DB("testdb").C("users").Insert(docs...)
	if err != nil {
		fmt.Println(err)
		return
	}
    //当url串中包含database名时，DB("")直接使用url串中的数据库
	db := client.DB("").C("users")
	var result []user
    //查询数据，带分页
	query := db.Find(map[string]any{"age": bson.M{"$gte": 25}})
	err = query.Skip(1).Limit(2).Sort("age").All(&result)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(result)
    //按之前的查询语句获取总记录数
	count, err := query.CountAll()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("总记录数：%d\n", count)
}

```

## 二、SDK详细介绍

### 2.1  公共函数

#### func Dial()函数

```go
func Dial(url string) (*Session, error)
```

根据mongodb的url连接串建立数据库连接，其中url串格式:

```
mongodb://[user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
```

options选项说明:

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



1.单机库连接串: *mongodb://user:password@localhost:27017/testdb?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256*

2.集群库连接串: *mongodb://user:password@server1:27017,server2:27017/testdb?retryWrites=true&serverSelectionTimeoutMS=5000&connectTimeoutMS=10000&authSource=test&authMechanism=SCRAM-SHA-256&replicaSet=repSet1*

#### func DialWithTimeout()函数

```go
func DialWithTimeout(url string, timeout time.Duration, poolMax int) (*Session, error)
```

带参数连接MongoDB数据库，参数:

| 参数名  | 参数说明                     |
| ------- | ---------------------------- |
| url     | MongoDB连接串                |
| timeout | 连接超时，最小超时10秒       |
| poolMax | 连接池最大值，默认最小值为10 |

#### func ToAnySlice()函数

```go
func ToAnySlice(docs any) []any
```

将结构对象数组转换成[]any函数

### 2.2  Session类

数据库连接会话类

#### func (s *Session) Clone()函数

```go
func (s *Session) Clone() *Session
```

克隆一个MongoDB连接。

#### func (s *Session) Copy() 函数

```go
func (s *Session) Copy() *Session
```

复制一个MongoDB连接，复制的连接是新建的连接。

#### func (s *Session) DB()函数 

```go
func (s *Session) DB(name string) *Database
```

选择连接的指定库名为name的数据库，返回*Database对象，相当于use命令。

若name参数为空字符串，则选择连接串中自带的数据库。

#### func (s *Session) Close()函数

```go
func (s *Session) Close()
```

关闭连接

#### func (s *Session) Ping()函数

```go
func (s *Session) Ping() error
```

检测数据库连接是否正常

#### func (s *Session) DatabaseNames()函数

```go
func (s *Session) DatabaseNames() (names []string, err error)
```

获取MongoDB中所有的数据库名称，相当于show dbs命令

### 2.3  Database类

数据库类，用于针对指定库名称的数据库的操作

#### func (d *Database) C(name string)函数

```go
func (d *Database) C(name string) *Collection
```

指定Collection表名称，获得表对象*Collection，若指定表尚未创建，则在插入第一条记录时自动创建

#### func (d *Database) Session()函数

```go
func (d *Database) Session() *Session
```

获取当前Database对象对应的连接会话对象

#### func (db *Database) CollectionNames()函数

```go
func (db *Database) CollectionNames() (names []string, err error)
```

获取当前数据库中所有的collection表名称

#### func (d *Database) AddUser() 函数

```go
func (d *Database) AddUser(username, password string, readOnly bool) error
```

添加数据库的用户，当前连接用户必须要有当前库的管理权限。新增的用户对本库只有只读和读写两种权限

| 参数     | 参数说明                        |
| -------- | ------------------------------- |
| username | 用户名                          |
| password | 登录密码                        |
| readOnly | 是否只读，true-只读，false-读写 |

#### func (d *Database) RemoveUser()函数

```go
func (d *Database) RemoveUser(username string) error
```

删除数据库的用户

#### func (d *Database) DropDatabase()函数

```go
func (d *Database) DropDatabase() error
```

删除当前数据库

### 2.4  Collection类

collection表对象，针对当前表的所有操作

#### func (c *Collection) Database()函数

```go
func (c *Collection) Database() *Database
```

获取当前表对应的数据库对象

#### func (c *Collection) DropCollection()函数

```go
func (c *Collection) DropCollection() error
```

删除当前表

#### func (c *Collection) Insert()函数

```go
func (c *Collection) Insert(docs ...any) error
```

插入数据，插入一条记录，也可以批量插入多条数据

#### func (c *Collection) Update()函数

```go
func (c *Collection) Update(selector, update interface{}) error
```

更新一条记录

| 参数     | 参数说明   |
| -------- | ---------- |
| selector | 查询条件   |
| update   | 更新的字段 |

#### func (c *Collection) Find()函数

```go
func (c *Collection) Find(query interface{}) *Query
```

查询语句返回*Query对象

#### func (c *Collection) UpdateAll() 函数

```go
func (c *Collection) UpdateAll(selector interface{}, update interface{}) (*ChangeInfo, error)
```

批量更新记录

#### func (c *Collection) UpdateId()函数

```go
func (c *Collection) UpdateId(id, update interface{}) error
```

通过id更新记录

#### func (c *Collection) Upsert()函数

```go
func (c *Collection) Upsert(selector, update interface{}) (*mongo.UpdateResult, error)
```

插入或更新函数

#### func (c *Collection) UpsertId()函数

```go
func (c *Collection) UpsertId(id, update interface{}) (*mongo.UpdateResult, error)
```

按ID插入或更新

#### func (c *Collection) Remove() 函数

```go
func (c *Collection) Remove(selector interface{}) error
```

按条件删除符合条件的第一条记录

#### func (c *Collection) RemoveId()函数

```go
func (c *Collection) RemoveId(id interface{}) error
```

按ID删除记录

#### func (c *Collection) RemoveAll()函数

```go
func (c *Collection) RemoveAll(selector interface{}) (*mongo.DeleteResult, error)
```

删除符合条件的所有记录

#### func (c *Collection) Count()函数

```go
func (c *Collection) Count() (int, error)
```

统计表的总记录数

#### func (c *Collection) EnsureIndex() 函数

```go
func (c *Collection) EnsureIndex(index Index) error
```

创建索引

### 2.5  Query类

专用于查询处理的链式处理

#### func (q *Query) Collection()函数

```go
func (q *Query) Collection() *Collection
```

返回上级表对象

#### func (q *Query) Select()函数

```go
func (q *Query) Select(selector interface{}) *Query
```

选择返回的字段

#### func (q *Query) Skip()函数

```go
func (q *Query) Skip(n int) *Query
```

跳过记录数，用于分页

#### func (q *Query) Limit()函数

```go
func (q *Query) Limit(n int) *Query
```

返回最大记录数，用于分页

#### func (q *Query) Sort()函数

```go
func (q *Query) Sort(fields ...string) *Query
```

按指定字段排序，排序优先于Skip和Limit

#### func (q *Query) One()函数

```go
func (q *Query) One(result interface{}) error
```

查询返回第一条记录

#### func (q *Query) All()函数

```go
func (q *Query) All(result interface{}) error
```

返回符合条件的所有记录，包括分页控制条件

#### func (q *Query) Count()函数

```go
func (q *Query) Count() (int, error)
```

统计符合条件的记录数，包括Skip和Limit控制

#### func (q *Query) CountAll()函数

```go
func (q *Query) CountAll() (int, error)
```

统计符合条件的总记录数，Skip和Limit不生效

