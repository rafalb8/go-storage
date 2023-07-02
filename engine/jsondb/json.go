package jsondb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rafalb8/go-maps"
	"github.com/rafalb8/go-maps/types"
	"github.com/rafalb8/go-storage"
	"github.com/rafalb8/go-storage/encoding"
	"github.com/rafalb8/go-storage/encoding/key"
	"github.com/rafalb8/go-storage/encoding/value"
	"github.com/rafalb8/go-storage/helpers"
	"github.com/rafalb8/go-storage/internal"
	"github.com/rafalb8/go-storage/internal/iter"
	"github.com/rafalb8/go-storage/options"
)

var (
	_ storage.Connection = (*JsonDB)(nil)
)

type JsonDB struct {
	data maps.EventfulMaper[string, []byte] // database data

	path       string // path to db
	singleFile bool

	ticker   *time.Ticker   // ticker for db file sync
	encoding encoding.Coder // db key/value encoder

	// prefix mutex map
	pfxMutex maps.Maper[string, sync.Locker]

	// cancel for data map event watcher/hub
	cancel context.CancelFunc

	// Logger
	lg storage.Logger
}

func New(opts ...JsonDBOpts) (storage.Connection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	j := &JsonDB{
		ticker:   time.NewTicker(time.Second),
		encoding: encoding.NewCoder(key.Simple, value.JSON),

		pfxMutex: maps.New[string, sync.Locker](nil).Safe(),
		cancel:   cancel,
		lg:       &internal.SimpleLogger{},
	}

	// Apply options
	for _, opt := range opts {
		err := opt(j)
		if err != nil {
			return nil, err
		}
	}

	var data map[string]any
	if internal.PathExists(j.path) {
		// Load db
		file, err := os.Open(j.path)
		if err != nil {
			return nil, err
		}

		err = json.NewDecoder(file).Decode(&data)
		if err != nil {
			return nil, err
		}
	}

	j.data = maps.New(iter.MapMap(data, func(v any) []byte {
		b, _ := json.Marshal(v)
		return b
	})).Eventful(ctx, 10)

	// Start save ticker
	go func() {
		for range j.ticker.C {
			err := j.Save()
			if err != nil {
				j.lg.Error(err)
			}
		}
	}()

	return j, nil
}

func (j *JsonDB) Encoding() encoding.Coder {
	return j.encoding
}

func (j *JsonDB) PrintDebug(pfx string) error {
	var err error
	out := map[string]any{}

	maps.NewBucket[[]byte](j.data, pfx).ForEach(func(k string, v []byte) error {
		out[k], err = helpers.Decode[any](j.encoding, v)
		if err != nil {
			return err
		}
		return nil
	})

	internal.PrintJSON(out)
	return err
}

func (j *JsonDB) Save() error {
	// unmarshal every value to map
	out := map[string]any{}
	j.data.ForEach(func(k string, v []byte) error {
		var val any
		err := json.Unmarshal(v, &val)
		out[k] = val
		return err
	})

	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(j.path, data, 0665)
	if err != nil {
		return err
	}
	return nil
}

func (j *JsonDB) Close() {
	j.cancel()
	j.ticker.Stop()
	err := j.Save()
	if err != nil {
		j.lg.Error(err)
	}
}

func (j *JsonDB) Bucket(bucket ...string) *storage.Bucket {
	return storage.NewBucket(j, j.encoding.DecodeBucket(bucket...)...)
}

func (j *JsonDB) Set(k string, v any, op ...storage.Option) error {
	j.lg.Debug("SET", k, v)
	data, err := j.encoding.EncodeValue(v)
	if err != nil {
		return err
	}
	j.data.Set(k, data)

	for _, opt := range op {
		switch opt := opt.(type) {
		case *options.TTLOption:
			go func(d time.Duration) {
				time.Sleep(d)
				j.data.Delete(k)
			}(opt.Value)

		default:
			j.lg.Warn("Unsupported option: %T", opt)
		}
	}

	return nil
}

func (j *JsonDB) Get(k string, v any) error {
	j.lg.Debug("GET", k)
	data, exists := j.data.GetFull(k)
	if !exists {
		return fmt.Errorf("get %s: %w", k, storage.ErrNotFound)
	}
	return j.encoding.DecodeValue(data, v)
}

func (j *JsonDB) Exists(k string) bool {
	j.lg.Debug("EXISTS", k)
	return j.data.Exists(k)
}

func (j *JsonDB) Delete(k string) error {
	j.lg.Debug("DELETE", k)
	j.data.Delete(k)
	return nil
}

func (j *JsonDB) Len(pfx string) (int, error) {
	j.lg.Debug("LEN", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Len(), nil
}

func (j *JsonDB) Keys(pfx string) ([]string, error) {
	j.lg.Debug("KEYS", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Keys(), nil
}

func (j *JsonDB) Values(pfx string) ([][]byte, error) {
	j.lg.Debug("VALUES", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Values(), nil
}

func (j *JsonDB) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	j.lg.Debug("ITER", pfx)
	out := make(chan types.Item[string, []byte])
	go func() {
		defer close(out)
		for item := range maps.NewBucket[[]byte](j.data, pfx).Iter() {
			out <- types.Item[string, []byte]{
				Key:   item.Key,
				Value: item.Value,
			}
		}
	}()
	return out
}

func (j *JsonDB) Watch(ctx context.Context, pfx string) types.Watcher[string, []byte] {
	j.lg.Debug("WATCH", pfx)
	out := make(chan types.WatchMsg[string, []byte])
	go func() {
		defer close(out)
		for event := range maps.NewBucket[[]byte](j.data, pfx).Watch(ctx) {
			out <- types.WatchMsg[string, []byte]{
				Event: event.Event,
				Item: types.Item[string, []byte]{
					Key: event.Key, Value: event.Value,
				},
			}
		}
	}()
	return out
}

func (j *JsonDB) Tx(pfx string, fn func(tx storage.Transactioner) error) error {
	j.lg.Debug("TX", pfx)

	var mtx sync.Locker
	j.pfxMutex.Commit(func(data map[string]sync.Locker) {
		var exists bool
		mtx, exists = data[pfx]
		if !exists {
			mtx = &sync.Mutex{}
			data[pfx] = mtx
		}
	})

	mtx.Lock()
	defer mtx.Unlock()
	return fn(j.Bucket(pfx))
}
