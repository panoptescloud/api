package bus

import (
	"fmt"
	"reflect"
)

type Command interface {
	GetName() string
}

type Query interface {
	GetName() string
}

type commandHandler[T Command] func(T) error
type queryHandler[T Query, TResult any] func(T) (TResult, error)

type CommandMiddleware func(next func(Command) error) func(Command) error
type QueryMiddleware func(next func(Query) (any, error)) func(Query) (any, error)

type Bus struct {
	registry          map[reflect.Type]any
	commandMiddleware []CommandMiddleware
	queryMiddleware   []QueryMiddleware
}

func AddCommandMiddleware(b *Bus, mw CommandMiddleware) {
	b.commandMiddleware = append(b.commandMiddleware, mw)
}

func AddQueryMiddleware(b *Bus, mw QueryMiddleware) {
	b.queryMiddleware = append(b.queryMiddleware, mw)
}

func RegisterCommand[T Command](b *Bus, h commandHandler[T], mws ...func(commandHandler[T]) commandHandler[T]) {
	typedHandler := h
	for i := len(mws) - 1; i >= 0; i-- {
		typedHandler = mws[i](typedHandler)
	}

	base := func(c Command) error {
		cmd := c.(T)
		return typedHandler(cmd)
	}

	wrapped := base
	for i := len(b.commandMiddleware) - 1; i >= 0; i-- {
		wrapped = b.commandMiddleware[i](wrapped)
	}

	var zero T
	typ := reflect.TypeOf(zero)
	b.registry[typ] = wrapped
}

func Dispatch[T Command](b *Bus, cmd T) error {
	typ := reflect.TypeOf(cmd)
	h, ok := b.registry[typ]
	if !ok {
		return fmt.Errorf("no handler registered for command %s", cmd.GetName())
	}
	return h.(func(Command) error)(cmd)
}

func RegisterQuery[T Query, TResult any](b *Bus, h queryHandler[T, TResult], mws ...func(queryHandler[T, TResult]) queryHandler[T, TResult]) {
	typedHandler := h
	for i := len(mws) - 1; i >= 0; i-- {
		typedHandler = mws[i](typedHandler)
	}

	base := func(q Query) (any, error) {
		query := q.(T)
		return typedHandler(query)
	}

	wrapped := base
	for i := len(b.queryMiddleware) - 1; i >= 0; i-- {
		wrapped = b.queryMiddleware[i](wrapped)
	}

	var zero T
	typ := reflect.TypeOf(zero)
	b.registry[typ] = wrapped
}

func RunQuery[T Query, TResult any](b *Bus, q T) (TResult, error) {
	typ := reflect.TypeOf(q)
	h, ok := b.registry[typ]
	if !ok {
		var zero TResult
		return zero, fmt.Errorf("no handler registered for query %s", q.GetName())
	}

	result, err := h.(func(Query) (any, error))(q)
	if err != nil {
		var zero TResult
		return zero, err
	}

	return result.(TResult), nil
}

func New() *Bus {
	return &Bus{
		registry: make(map[reflect.Type]any),
	}
}
