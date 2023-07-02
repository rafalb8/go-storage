package memory

import (
	"context"
	"fmt"
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
	"github.com/rafalb8/go-storage/options"
)

var (
	_ storage.Connection = (*InMemory)(nil)
)

type InMemory struct {
	data     maps.EventfulMaper[string, []byte] // database data
	encoding encoding.Coder                     // db key/value encoder

	// prefix mutex map
	pfxMutex maps.Maper[string, sync.Locker]

	// cancel for data map event watcher/hub
	cancel context.CancelFunc

	// Logger
	lg storage.Logger
}

func New(opts ...MemoryOpts) (storage.Connection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := &InMemory{
		data:     maps.New[string, []byte](nil).Eventful(ctx, 10),
		encoding: encoding.NewCoder(key.Binary, value.CBOR),

		pfxMutex: maps.New[string, sync.Locker](nil).Safe(),
		cancel:   cancel,
		lg:       &internal.SimpleLogger{},
	}

	// Apply options
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *InMemory) Close() {
	m.cancel()
}

func (m *InMemory) Encoding() encoding.Coder {
	return m.encoding
}

func (m *InMemory) PrintDebug(pfx string) error {
	var err error
	out := map[string]any{}

	maps.NewBucket[[]byte](m.data, pfx).ForEach(func(k string, v []byte) error {
		out[k], err = helpers.Decode[any](m.encoding, v)
		if err != nil {
			return err
		}
		return nil
	})

	internal.PrintJSON(out)
	return err
}

func (m *InMemory) Bucket(bucket ...string) *storage.Bucket {
	return storage.NewBucket(m, m.encoding.DecodeBucket(bucket...)...)
}

func (m *InMemory) Set(k string, v any, op ...storage.Option) error {
	m.lg.Debug("SET", k, v)
	data, err := m.encoding.EncodeValue(v)
	if err != nil {
		return err
	}
	m.data.Set(k, data)

	for _, opt := range op {
		switch opt := opt.(type) {
		case *options.TTLOption:
			go func(d time.Duration) {
				time.Sleep(d)
				m.data.Delete(k)
			}(opt.Value)

		default:
			m.lg.Warn("Unsupported option: %T", opt)
		}
	}
	return nil
}

func (m *InMemory) Get(k string, v any) error {
	m.lg.Debug("GET", k)
	data, exists := m.data.GetFull(k)
	if !exists {
		return fmt.Errorf("get %s: %w", k, storage.ErrNotFound)
	}
	return m.encoding.DecodeValue(data, v)
}

func (m *InMemory) Exists(k string) bool {
	m.lg.Debug("EXISTS", k)
	return m.data.Exists(k)
}

func (m *InMemory) Delete(k string) error {
	m.lg.Debug("DELETE", k)
	m.data.Delete(k)
	return nil
}

func (m *InMemory) Len(pfx string) (int, error) {
	m.lg.Debug("LEN", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Len(), nil
}

func (m *InMemory) Keys(pfx string) ([]string, error) {
	m.lg.Debug("KEYS", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Keys(), nil
}

func (m *InMemory) Values(pfx string) ([][]byte, error) {
	m.lg.Debug("VALUES", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Values(), nil
}

func (m *InMemory) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	m.lg.Debug("ITER", pfx)
	out := make(chan types.Item[string, []byte])
	go func() {
		defer close(out)
		for item := range maps.NewBucket[[]byte](m.data, pfx).Iter() {
			out <- types.Item[string, []byte]{
				Key:   item.Key,
				Value: item.Value,
			}
		}
	}()
	return out
}

func (m *InMemory) Watch(ctx context.Context, pfx string) types.Watcher[string, []byte] {
	m.lg.Debug("WATCH", pfx)
	out := make(chan types.WatchMsg[string, []byte])
	go func() {
		defer close(out)
		for event := range maps.NewBucket[[]byte](m.data, pfx).Watch(ctx) {
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

func (m *InMemory) Tx(pfx string, fn func(tx storage.Transactioner) error) error {
	m.lg.Debug("TX", pfx)

	var mtx sync.Locker
	m.pfxMutex.Commit(func(data map[string]sync.Locker) {
		var exists bool
		mtx, exists = data[pfx]
		if !exists {
			mtx = &sync.Mutex{}
			data[pfx] = mtx
		}
	})

	mtx.Lock()
	defer mtx.Unlock()
	return fn(m.Bucket(pfx))
}
