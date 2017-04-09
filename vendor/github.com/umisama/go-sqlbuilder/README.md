# umisama/go-sqlbuilder
**go-sqlbuilder** is a SQL-query builder for golang.  This supports you using relational database with more readable and flexible code than raw SQL query string.

[![Build Status](https://travis-ci.org/umisama/go-sqlbuilder.svg?branch=master)](https://travis-ci.org/umisama/go-sqlbuilder)
[![Coverage Status](https://coveralls.io/repos/umisama/go-sqlbuilder/badge.svg)](https://coveralls.io/r/umisama/go-sqlbuilder)

## Support
 * Generate SQL query programmatically.
   * fluent flexibility! yeah!!
 * Basic SQL statements
   * SELECT/INSERT/UPDATE/DELETE/DROP/CREATE TABLE/CREATE INDEX
 * Strict error checking
 * Some database server
   * Sqlite3([mattn/go-sqlite3](https://github.com/mattn/go-sqlite3))
   * MySQL([ziutek/mymysql](https://github.com/ziutek/mymysql))
   * MySQL([go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))
   * PostgresSQL([lib/pq](https://github.com/lib/pq))
 * Subquery in SELECT FROM clause

## TODO
 * Support UNION clause
 * Support LOCK clause

## Quick usage

```go
import (
	sb "github.com/umisama/go-sqlbuilder"
	"github.com/umisama/go-sqlbuilder/dialects"
)

db, err := sql.Open("sqlite3", ":memory:")
if err != nil {
	fmt.Println(err.Error())
	return
}

// Set dialect first
// dialects are in github.com/umisama/go-sqlbuilder/dialects
sb.SetDialect(TestDialect{})

// Define a table
tbl_person := sb.NewTable(
	"PERSON",
	&sb.TableOption{},
	sb.IntColumn("id", &sb.ColumnOption{
		PrimaryKey: true,
	}),
	sb.StringColumn("name", &sb.ColumnOption{
		Unique:  true,
		Size:    255,
		Default: "no_name",
	}),
	sb.DateColumn("birth", nil),
)

// Create Table
query, args, err := sb.CreateTable(tbl_person).ToSql()
if err != nil {
	fmt.Println(err.Error())
	return
}
_, err = db.Exec(query, args...)
if err != nil {
	fmt.Println(err.Error())
	return
}

// Insert data
// (Table).C function returns a column object.
query, args, err = sb.Insert(tbl_person).
	Set(tbl_person.C("name"), "Kurisu Makise").
	Set(tbl_person.C("birth"), time.Date(1992, time.July, 25, 0, 0, 0, 0, time.UTC)).
	ToSql()
_, err = db.Exec(query, args...)
if err != nil {
	fmt.Println(err.Error())
	return
}

// Query
// (Column).Eq returns a condition object for equal(=) operator.  See
var birth time.Time
query, args, err = sb.Select(tbl_person).Columns(
	tbl_person.C("birth"),
).Where(
	tbl_person.C("name").Eq("Kurisu Makise"),
).ToSql()
err = db.QueryRow(query, args...).Scan(&birth)
if err != nil {
	fmt.Println(err.Error())
	return
}

fmt.Printf("Kurisu's birthday is %s,%d %d", birth.Month().String(), birth.Day(), birth.Year())

// Output:
// Kurisu's birthday is July,25 1992
```

## Examples
### Initialize
off course, go getable.

```shell-script
$ go get github.com/umisama/go-sqlbuilder
```

I recomended to set "sb" as sqlbuilder's shorthand.

```go
import sb "github.com/umisama/go-sqlbuilder"

// First, you set dialect for your DB
func init (
	sb.SetDialect(sb.SqliteDialect{})
)
```

### Define a table
Sqlbuilder needs table definition to strict query generating.  Any statement checks column type and constraints.

```go
tbl_person := sb.NewTable(
	"PERSON",
	&sb.TableOption{},
	sb.IntColumn("id", &sb.ColumnOption{
		PrimaryKey: true,
	}),
	sb.StringColumn("name", &sb.ColumnOption{
		Unique:  true,
		Size:    255,
		Default: "no_name",
	}),
	sb.DateColumn("birth", nil),
)
```

#### Table Options
##### Unique [][]string
Sets UNIQUE options to table.

example:

```go
&sb.TableOption{
	Unique: [][]string{
		{"hoge", "piyo"},
		{"fuga"},
	}
}
```

```
CREATE TABLE PERSON ( "id" integer, ~~~, UNIQUE("hoge", "piyo"), UNIQUE("fuga"))
```

#### Column Options
##### PrimaryKey bool
`true` for add primary key option.

##### NotNull bool
`true` for add UNIQUE option.

##### Unique bool
`true` for add UNIQUE option to column.

example:

```go
IntColumn("test", &sb.ColumnOption{
	Unique: true,
})
```

```
"test" INTEGER UNIQUE
```

##### AutoIncrement bool
`true` for add AutoIncrement option to column.

##### Size int
Sets size for string column.
example:

```go
StringColumn("test", &sb.ColumnOption{
	Size: 255,
})
```

```
"test" VARCHAR(255)
```

##### SqlType string
Sets type for column on AnyColumn.

```go
AnyColumn("test", &sb.ColumnOption{
	ColumnType: "BLOB",
})
```

```
"test" BLOB
```

##### Default interface{}
Sets default value.  Default's type need to be same as column.


```go
StringColumn("test", &sb.ColumnOption{
	Size:    255,
	Default: "empty"
})
```

```
"test" VARCHAR(255) DEFAILT "empty"
```

### CRATE TABLE statement
Sqlbuilder has a `Statement` object generating CREATE TABLE statement from table object.  
`Statement` objects have `ToSql()` method. it returns query(string), placeholder arguments([]interface{}) and error.

```go
query, args, err := sb.CreateTable(tbl_person).ToSql()
if err != nil {
	panic(err)
}
// query == `CREATE TABLE "PERSON" ( "id" INTEGER PRIMARY KEY, "value" INTEGER );`
// args  == []interface{}{}
// err   == nil
```

You can exec with ```database/sql``` package or Table-struct mapper(for example, gorp).  
here is example,

```go
db, err := sql.Open("sqlite3", ":memory:")
if err != nil {
	panic(err)
}
_, err = db.Exec(query, args...)
if err != nil {
	panic(err)
}
```

### INSERT statement
Sqlbuilder can generate INSERT statement.  You can checkout a column with `Table.C([column_name])` method.

```go
query, args, err := sb.Insert(table1).
	Columns(table1.C("id"), table1.C("value")).
	Values(1, 10).
	ToSql()
// query == `INSERT INTO "TABLE_A" ( "id", "value" ) VALUES ( ?, ? );`
// args  == []interface{}{1, 10}
// err   == nil
```

Or,  can use `Set()` method.

```go
query, args, err := sb.Insert(table1).
	Set(table1.C("id"), 1).
	Set(table1.C("value"), 10).
	ToSql()
// query == `INSERT INTO "TABLE_A" ( "id", "value" ) VALUES ( ?, ? );`
// args  == []interface{}{1, 10}
// err   == nil
```

### SELECT statement
Sqlbuilder can generate SELECT statement with readable interfaces.  Condition object is generated from column object.

```go
query, args, err := sb.Select(table1.C("id"), table1.C("value")).
	From(table1).
	Where(
		table1.C("id").Eq(10),
	).
	Limit(1).OrderBy(false, table1.C("id")).
	ToSql()
// query == `SELECT "TABLE_A"."id", "TABLE_A"."value" FROM "TABLE_A" WHERE "TABLE_A"."id"=? ORDER BY "TABLE_A"."id" ASC LIMIT ?;`
// args  == []interface{}{10, 1}
// err   == nil
```

See [godoc.org](http://godoc.org/github.com/umisama/go-sqlbuilder#SelectStatement) for more options

### Condition
You can define condition with Condition objects.  Condition object create from ```Column```'s method.

| example operation                     |  output example              |
|:-------------------------------------:|:--------------------------:|
|```table1.C("id").Eq(10)```              | "TABLE1"."id"=10           |
|```table1.C("id").Eq(table2.C("id"))```    | "TABLE1"."id"="TABLE2"."id"|

More than one condition can combine with AND & OR operator.

| example operation                     |  output example              |
|:-------------------------------------:|:--------------------------:|
|```And(table1.C("id").Eq(1), table2.C("id").Eq(2)``` | "TABLE1"."id"=1 AND "TABLE2"."id"=1 |
|```Or(table1.C("id").Eq(1), table2.C("id").Eq(2)```  | "TABLE1"."id"=1 OR "TABLE2"."id"=1 |

Sqlbuilder is supporting most common condition operators.  
Here is supporting:

| columns method        | means                  | SQL operator  |      example         |
|:---------------------:|:----------------------:|:-------------:|:--------------------:|
|Eq(Column or value)    |EQUAL TO                |    ```=```    | "TABLE"."id" = 10    |
|NotEq(Column or value) |NOT EQUAL TO            |   ```<>```    | "TABLE"."id" <> 10   |
|Gt(Column or value)    |GRATER-THAN             |    ```>```    | "TABLE"."id" > 10    |
|GtEq(Column or value)  |GRATER-THAN OR EQUAL TO |   ```>=```    | "TABLE"."id" >= 10   |
|Lt(Column or value)    |LESS-THAN               |    ```<```    | "TABLE"."id" < 10    |
|LtEq(Column or value)  |LESS-THAN OR EQUAL TO   |   ```<=```    | "TABLE"."id" <= 10   |
|Like(string)           |LIKE                    |  ```LIKE```   | "TABLE"."id" LIKE "%hoge%"   |
|In(values array)       |IN                      |   ```IN```    | "TABLE"."id" IN ( 1, 2, 3 ) |
|NotIn(values array)    |NOT IN                  | ```NOT IN```  | "TABLE"."id" NOT IN ( 1, 2, 3 ) |
|Between(loewer, higher int) |BETWEEN            | ```BETWEEN``` | "TABLE"."id" BETWEEN 10 AND 20)|

Document for all: [godoc(Column)](http://godoc.org/github.com/umisama/go-sqlbuilder#Column)

## More documents
[godoc.org](http://godoc.org/github.com/umisama/go-sqlbuilder)

## License
under the MIT license
