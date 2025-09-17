package main

import (
	"fmt"
	"reflect"
)

// --- base interfaces

type Command interface {
	GetName() string
}

type Query interface {
	GetName() string
}

// --- handler types

type commandHandler[T Command] func(T) error
type queryHandler[T Query, TResult any] func(T) (TResult, error)

// --- middleware types

type CommandMiddleware func(next func(Command) error) func(Command) error
type QueryMiddleware func(next func(Query) (any, error)) func(Query) (any, error)

// --- bus

type Bus struct {
	registry          map[reflect.Type]any
	commandMiddleware []CommandMiddleware
	queryMiddleware   []QueryMiddleware
}

func NewBus() *Bus {
	return &Bus{registry: make(map[reflect.Type]any)}
}

// Global middleware
func UseCommand(b *Bus, mw CommandMiddleware) {
	b.commandMiddleware = append(b.commandMiddleware, mw)
}

func UseQuery(b *Bus, mw QueryMiddleware) {
	b.queryMiddleware = append(b.queryMiddleware, mw)
}

// Register command handler
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

// Register query handler
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

// --- commands

type CreateUserCommand struct {
	Name string
}

func (CreateUserCommand) GetName() string { return "users.create" }

// --- queries

type GetUserQuery struct {
	ID string
}

func (GetUserQuery) GetName() string { return "users.get" }

// --- services

type UsersService struct{}

func (us *UsersService) CreateUser(cmd CreateUserCommand) error {
	fmt.Printf("Creating user: %s\n", cmd.Name)
	return nil
}

func (us *UsersService) GetUser(q GetUserQuery) (string, error) {
	return "User(" + q.ID + ")", nil
}

// --- middleware

func LoggingCommand(next func(Command) error) func(Command) error {
	return func(c Command) error {
		fmt.Printf("[CMD-LOG] %s\n", c.GetName())
		return next(c)
	}
}

func LoggingQuery(next func(Query) (any, error)) func(Query) (any, error) {
	return func(q Query) (any, error) {
		fmt.Printf("[QRY-LOG] %s\n", q.GetName())
		return next(q)
	}
}

// --- main

func main() {
	bus := NewBus()
	usersSvc := &UsersService{}

	// Global middleware
	UseCommand(bus, LoggingCommand)
	UseQuery(bus, LoggingQuery)

	// Register handlers
	RegisterCommand(bus, usersSvc.CreateUser)
	RegisterQuery(bus, usersSvc.GetUser)

	// Dispatch command
	Dispatch(bus, CreateUserCommand{Name: "Alice"})

	// Dispatch query
	user, _ := RunQuery[GetUserQuery, string](bus, GetUserQuery{ID: "123"})
	fmt.Println("Query result:", user)
}
