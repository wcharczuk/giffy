Spiffy
======

This is a very bare bones database interface for golang. It abstracts away a bunch of boilerplate so that the developer can concentrate on writing their app.

It does not abstract away actual sql, however. 

# Gotchas & General Notes #

There is a standing pattern that every action (query, exec, create, update etc.) has a corresponding ...InTx method that these top level methods actually call into with `nil` as the tx. If the tx is nil, a direct connection, free of a wrapping transaction will be created for that command during the `prepare` phase of the command execution. 

# Mapping Structs Using Spiffy #

A sample database mapped type:
```go
type MyTable struct {
	Id int `db:"id,serial,pk"`
	Name string
	Excluded `db:"-"`
}
func (mt MyTable) TableName() string {
	return "my_table"
}
```

Two important things: we define `tags` on struct members to tell the orm how to interact with the db. We also define a method that returns
the table name for the struct as it is mapped in the db. 

Tags are laid out in the following format `db:"<column_name>,<options>,..."`, where after the `column_name` there can be multiple `options`. An example above is `id,serial,pk`, which translates to a column name `id` and options `serial,pk` respectively. 

Options include:
- `serial` : denotes a column that will be read back on `Create` (there can only be 1 at this time)
- `pk` : deontes a column that consitutes a primary key. Will be used when creating SQL where clauses.
- `readonly` : denotes a column that is only read, not written to the db.

# Managing Connections and Aliases #

The next step in running a database driven app is to tell the app how to connect to the db. There are 4 required pieces of info to do this: `host`, `db name`, `username`, `password`. Note: `host` should include the port if it's non-standard. `db name` is the database you're hitting. 

We can manage connections (and save a default so it doesn't need to be passed around) with "Aliases".

*Example:*
```golang
connection := spiffy.NewConnection("localhost", "my_db", "postgres", "super_secret_pw"))
spiffy.InitDefault(connection)
```

The above snipped creates a connection, and then saves it as the default connection. This lets us then call `spiffy.DB()` to retrieve this connection. Alternatively we could spin up a connection and pass it around the app as pointer, but this get's tricky and it's easier just to save it to the a central location.

# Querying, Execing, Getting Objects from the Database #

There are two paradigms for interacting with the database; functions that return QueryResults, and functions that just return errors. 

## Execing ##

Simple execute operations can be done with `Exec` or `ExecInTx` functions. 

*Example:*
```golang
err := spiffy.DB().Exec("delete from my_table where id = $1", obj_id)
```

When we need to pass parameters to the queries, use `$1` numbered tokens to denote the parameter in the sql. We then need to pass that parameter as an argument to `Exec` in the order that maps to the numbered token.

## Querying ###

Querying in Spiffy can be done with the `Query` or `QueryInTx` functions. Each takes SQL as it's main parameter. That's it, no complicated DSL's for replacing sql, just write it yourself. 

*Struct Output Example*
```golang
obj := MyObject{}
err := spiffy.DB().Query("select * from my_table where id = $1", obj_id).Out(&obj)
```

*Slice Ouptut Example:*
```golang
objs := []MyObject{}
err := spiffy.DB().Query("select * from my_table").OutMany(&objs)
```

In order to query the database, we need a query and a target for the output. The output can be a single struct, or a slice of structs. Which we're using determines if we use `Out` or `OutMany`. Like `Exec`, when we need to pass parameters to the queries, use `$1` numbered tokens to denote the parameter in the sql. We then need to pass that parameter as an argument to `Query` in the order that maps to the numbered token.

# CrUD Operations #

You can perform the following CrUD operations:
- `Create` or `CreateInTx` : create objects

*Example:*
```golang
obj := MyObj{...}
create_err := spiffy.DB().Create(&obj) //note the reference! this is incase we have to write back a serial id.
```

- `Update` or `UpdateInTx` : update objects

*Example:*
```golang
obj := MyObj{..}
err := spiffy.DB().GetById(&obj, objID) //note, there can be multiple params (!!) if there are multiple pks
obj.Property = "new_value"
err = spiffy.DB().Update(obj) //note we don't need a reference for this, as it's read only.
```

- `Delete` or `DeleteInTx` : delete objects

*Example:*
```golang
obj := MyObj{...}
err := spiffy.DB().GetById(&obj, objID) //note, there can be multiple params (!!) if there are multiple pks
err = spiffy.DB().Delete(obj) //note we don't need a reference for this, as it's read only.
```

# Performance #

Generally it's pretty good. There is a comparison test in `spiffy_test.go` if you want to see for yourself. It creates 5000 objects with 5 properties each, then reads them out using the orm or manual scanning.

The results were:

| manual  |   orm    |
|---------|----------|
|17.11ms  | 38.08ms  |

This would be shabby in .net land, but for Go where we can't dynamically emit populate methods, and we're stuck doing in-order population or by name population, it's ok.

If you implement `Populatable`, performance improves dramatically.

| manual  |   orm (Populatable)    |
|---------|------------------------|
|14.33ms  | 16.95ms                |

The strategy then is to impelement populate on your "hot read" objects, and let the orm figure out the other ones.
