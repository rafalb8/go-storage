# go-storage

## Supported engines

 - Memory
 - Json file
 - Etcd

## Usage

```go
import "github.com/rafalb8/go-storage/jsondb"

func connect() {
    db, err := jsondb.New("/path/to/file.json")
    // handle err ...
    db.Set("key", "value")
    fmt.Println(db.Get("key"))
}
```
