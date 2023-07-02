package helpers

import (
	"context"
	"fmt"

	"github.com/rafalb8/go-maps/types"
	"github.com/rafalb8/go-storage"
	"github.com/rafalb8/go-storage/encoding"
)

type WatchHelper interface {
	storage.Watcher
	Encoding() encoding.Coder
}

func Watch[T any](ctx context.Context, tx WatchHelper, key string) types.Watcher[string, T] {
	out := make(chan types.WatchMsg[string, T], 10)

	go func() {
		defer close(out)

		for event := range tx.Watch(ctx, key) {
			value, err := Decode[T](tx.Encoding(), event.Value)
			if err != nil {
				out <- types.WatchMsg[string, T]{
					Event: types.ErrorEvent,
					Item: types.Item[string, T]{
						Key: fmt.Sprintf("error decoding %s, %s", event.Key, err),
					},
				}
			}

			// Repackage event
			out <- types.WatchMsg[string, T]{
				Event: event.Event,
				Item: types.Item[string, T]{
					Key:   event.Key,
					Value: value,
				},
			}
		}
	}()

	return out
}

func Iter[T any](ctx context.Context, tx storage.Transactioner) <-chan types.Item[string, T] {
	out := make(chan types.Item[string, T], 10)

	go func() {
		defer close(out)

		for event := range tx.Iter(ctx, "") {
			value, err := Decode[T](tx.Encoding(), event.Value)
			if err != nil {
				out <- types.Item[string, T]{Key: fmt.Sprintf("error decoding %s, %s", event.Key, err)}
			}
			out <- types.Item[string, T]{Key: event.Key, Value: value}
		}
	}()

	return out
}

func Map[T any](tx storage.Bucketer) map[string]T {
	out := map[string]T{}
	for i := range Iter[T](context.Background(), tx) {
		out[i.Key] = i.Value
	}
	return out
}

func Values[T any](tx storage.Bucketer) ([]T, error) {
	vals, err := tx.Values()
	if err != nil {
		return nil, err
	}

	out := make([]T, len(vals))

	for i, data := range vals {
		value, err := Decode[T](tx.Encoding(), data)
		if err != nil {
			return nil, err
		}
		out[i] = value
	}

	return out, nil
}

func Get[T any](tx storage.Transactioner, key string) (T, error) {
	val := new(T)
	err := tx.Get(key, val)
	if err != nil {
		return *val, err
	}
	return *val, nil
}

func Decode[T any](c encoding.ValueCoder, data []byte) (T, error) {
	val := new(T)
	err := c.DecodeValue(data, val)
	return *val, err
}
