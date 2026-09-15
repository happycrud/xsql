# xsql — Go 语言类型安全 SQL 构建器与 DB 客户端

**xsql** 是驱动 [crud](https://github.com/goflower-io/crud) ORM 的 SQL 构建器和数据库客户端库。基于 Go 泛型提供流式 API，支持 MySQL、PostgreSQL 和 SQLite3，内置读写分离和操作级超时控制。

[English](README.md) | [crud](https://github.com/goflower-io/crud) | [golib](https://github.com/goflower-io/golib) | [示例代码](https://github.com/goflower-io/example)

---

## 核心特性

- **泛型字段操作** — `FieldOp[int64]`、`StrFieldOp` — 完全类型安全，无需类型断言
- **流式 SQL 构建** — SELECT / INSERT / UPDATE / DELETE 链式调用 API
- **多方言支持** — MySQL、PostgreSQL、SQLite3
- **读写分离** — 自动主从路由，读请求轮询负载均衡
- **操作级超时** — `QueryTimeout` 和 `ExecTimeout` 独立配置
- **可组合谓词** — `And`、`Or`、`Not` 组合复杂 WHERE 条件
- **自定义结果扫描** — 通过 JSON tag 将结果扫描到任意结构体
- **调试模式** — 一行代码打印生成的 SQL 和参数
- **可配置日志** — 接入任意 `Printf` 日志器，或复用 `log/slog`，支持运行时开启/关闭

---

## 安装

```bash
go get github.com/goflower-io/xsql
```

---

## 建立连接

```go
import (
    "github.com/goflower-io/xsql"
    "github.com/goflower-io/xsql/mysql"
)

db, err := mysql.NewDB(&xsql.Config{
    DSN:          "root:123456@tcp(127.0.0.1:3306)/test?parseTime=true&loc=Local",
    ReadDSN:      []string{"root:123456@tcp(127.0.0.1:3307)/test?parseTime=true&loc=Local"},
    Active:       20,
    Idle:         20,
    IdleTimeout:  24 * time.Hour,
    QueryTimeout: 10 * time.Second,  // 作用于 SELECT
    ExecTimeout:  10 * time.Second,  // 作用于 INSERT/UPDATE/DELETE
})
```

PostgreSQL 和 SQLite3：

```go
import "github.com/goflower-io/xsql/postgres"
db, _ := postgres.NewDB(&xsql.Config{
    DSN: "host=127.0.0.1 user=postgres password=123456 dbname=test sslmode=disable",
})

import "github.com/goflower-io/xsql/sqlite"
db, _ := sqlite.NewDB(&xsql.Config{
    DSN: "/path/to/my.db",
})
```

SQLite 提供三个可互换的驱动。默认的 `github.com/goflower-io/xsql/sqlite`
包使用 [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3)
（纯 Go，无需 cgo）。如需其他驱动，导入对应的子包：

```go
import "github.com/goflower-io/xsql/sqlite"         // 默认：ncruces/go-sqlite3（无需 cgo）
import "github.com/goflower-io/xsql/sqlite/mattn"   // mattn/go-sqlite3（cgo）
import "github.com/goflower-io/xsql/sqlite/modernc" // modernc.org/sqlite（纯 Go）
```

---

## 字段操作

xsql 通常通过 `crud` 生成的代码来使用。生成代码为每个字段声明类型化的操作符：

```go
// 由 crud 生成于 crud/user/user.go
var IdOp   = xsql.FieldOp[int64]("id")
var NameOp = xsql.StrFieldOp("name")
var AgeOp  = xsql.FieldOp[int64]("age")
```

### 数值 / 可比较字段 — `FieldOp[T]`

```go
IdOp.EQ(1)           // id = 1
IdOp.NEQ(1)          // id != 1
IdOp.GT(10)          // id > 10
IdOp.GTE(10)         // id >= 10
IdOp.LT(100)         // id < 100
IdOp.LTE(100)        // id <= 100
IdOp.In(1, 2, 3)     // id IN (1, 2, 3)
IdOp.NotIn(4, 5)     // id NOT IN (4, 5)
```

### 字符串字段 — `StrFieldOp`

```go
NameOp.EQ("alice")           // name = 'alice'
NameOp.NEQ("alice")          // name != 'alice'
NameOp.Contains("ali")       // name LIKE '%ali%'
NameOp.HasPrefix("ali")      // name LIKE 'ali%'
NameOp.HasSuffix("ice")      // name LIKE '%ice'
NameOp.IsNull()              // name IS NULL
NameOp.NotNull()             // name IS NOT NULL
```

### 组合谓词

```go
// 隐式 AND — 向 Where() 传多个参数
finder.Where(IdOp.GT(10), AgeOp.LT(30))
// WHERE id > 10 AND age < 30

// OR
finder.Where(xsql.OrOp(IdOp.EQ(1), IdOp.EQ(2)))
// WHERE id = 1 OR id = 2

// NOT
finder.Where(xsql.NotOp(AgeOp.In(10, 20)))
// WHERE NOT age IN (10, 20)

// 嵌套：(id > 10 OR age IN (18, 25)) AND name LIKE 'ali%'
finder.Where(
    xsql.OrOp(IdOp.GT(10), AgeOp.In(18, 25)),
    NameOp.HasPrefix("ali"),
)
```

---

## 聚合与表达式辅助函数

```go
xsql.Count("*")              // COUNT(*)
xsql.As(expr, "alias")       // expr AS alias
xsql.GT("cnt", 1)            // cnt > 1  — 用于 HAVING 子句
xsql.And(ps...)              // 用 AND 组合多个 *Predicate
xsql.Or(ps...)               // 用 OR 组合多个 *Predicate
xsql.GenP(col, op, val)      // 从动态字符串输入构建 *Predicate
```

### 从用户输入动态生成谓词

```go
// 支持的操作符字符串：EQ NEQ GT GTE LT LTE IN NOT_IN CONTAINS HAS_PREFIX HAS_SUFFIX
p, err := xsql.GenP("age", "GT", "18")
if err != nil {
    return err
}
finder.WhereP(p)
```

---

## 配置项说明

```go
type Config struct {
    DSN          string        // 写库（主库）连接字符串
    ReadDSN      []string      // 读库（从库）连接字符串列表，轮询负载均衡
    Active       int           // 最大打开连接数
    Idle         int           // 最大空闲连接数
    IdleTimeout  time.Duration // 连接最大空闲时间
    QueryTimeout time.Duration // SELECT 查询的 Context 超时时间
    ExecTimeout  time.Duration // INSERT/UPDATE/DELETE 的 Context 超时时间
}
```

---

## 日志配置

SQL 日志可选且完全可配置。通过 `SetLogger` 给 `*DB` 设置日志器，即可记录它执行的所有 SQL：

```go
import (
    "log/slog"
    "github.com/goflower-io/xsql"
)

db, _ := mysql.NewDB(&xsql.Config{DSN: "..."})
db.SetLogger(xsql.SlogLogger(slog.Default())) // 接入应用的 slog 日志器
```

`Logger` 只包含一个 `Printf` 方法，任何日志器都可以使用，包括标准库的
`*log.Logger`：

```go
type Logger interface {
    Printf(format string, v ...any)
}
```

SQL 的 context 和执行错误都会包含在传给 `Printf` 的消息中，日志器无需额外
方法。`xsql.SlogLogger` 可将 `*slog.Logger` 适配为 `Logger`，以 debug 级别输出。

也可以在运行时修改日志器，传 `nil` 关闭日志：

```go
db.SetLogger(myLogger) // 开启 / 替换
db.SetLogger(nil)      // 关闭
```

---

## 调试模式

`xsql.Debug` 可包装任意 `DBI`，打印每条 SQL（包括事务内执行的 SQL）。可通过
`WithLogger` 覆盖默认日志器（`log.Default()`）：

```go
// 包装任意 *DB，打印 SQL 和参数
user.Create(xsql.Debug(db)).SetUser(u).Save(ctx)
// [xsql] INSERT INTO `user` (`name`, `age`, `ctime`, `mtime`) VALUES (?, ?, ?, ?) [alice 18 ...]

// 自定义日志器
user.Create(xsql.Debug(db, xsql.WithLogger(xsql.SlogLogger(slog.Default())))).SetUser(u).Save(ctx)
```

---

## 读写路由规则

- **写操作**（`INSERT`、`UPDATE`、`DELETE`）始终路由到 `DSN`（主库）。
- **读操作**（`SELECT`）在 `ReadDSN` 从库列表中轮询。
- 若 `ReadDSN` 为空，读操作也使用 `DSN`。

---

## 数据库方言支持

| 方言 | 驱动 | 导入路径 |
|---|---|---|
| MySQL / MariaDB | `github.com/go-sql-driver/mysql` | `github.com/goflower-io/xsql/mysql` |
| PostgreSQL | `github.com/jackc/pgx/v5` | `github.com/goflower-io/xsql/postgres` |
| SQLite3（默认） | `github.com/ncruces/go-sqlite3` | `github.com/goflower-io/xsql/sqlite` |
| SQLite3（cgo） | `github.com/mattn/go-sqlite3` | `github.com/goflower-io/xsql/sqlite/mattn` |
| SQLite3（纯 Go） | `modernc.org/sqlite` | `github.com/goflower-io/xsql/sqlite/modernc` |

---

## 相关仓库

- [crud](https://github.com/goflower-io/crud) — 以 xsql 为运行时，从 SQL DDL 生成类型安全的 CRUD 代码
- [golib](https://github.com/goflower-io/golib) — HTTP/gRPC 应用服务器框架
- [example](https://github.com/goflower-io/example) — 展示四个库协同工作的全栈示例
