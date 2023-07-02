# go-storage

## Supported engines

 - Memory
 - Json file
 - Etcd

## Usage

```go
import "github.com/rafalb8/go-storage/engine/jsondb"

func connect() {
    db, err := jsondb.New(jsondb.File("/path/to/file.json"))
    // handle err ...

    // Set key in db
    db.Set("key", "value")

    val, err := helpers.Get[string](db, "key")
    // handle db err ...

    // print value
    fmt.Println(val)
}
```

## Planned features

 - [ ] JsonDB in multiple files