package test

import (
	"google.golang.org/protobuf/proto"

	"github.com/nhatthm/grpcmock/internal/grpctest"
)

// ItemBuilder builds an item.
type ItemBuilder struct {
	item *grpctest.Item
}

// From creates a new item from a source.
func (b ItemBuilder) From(src *grpctest.Item) ItemBuilder {
	b.item = clone(src)

	return b
}

// New creates a new item.
func (b ItemBuilder) New() *grpctest.Item {
	return clone(b.item)
}

// WithID sets the id.
func (b ItemBuilder) WithID(id int32) ItemBuilder {
	b.item = b.New()
	b.item.Id = id

	return b
}

// WithLocale sets the locale.
func (b ItemBuilder) WithLocale(locale string) ItemBuilder {
	b.item = b.New()
	b.item.Locale = locale

	return b
}

// WithName sets the name.
func (b ItemBuilder) WithName(name string) ItemBuilder {
	b.item = b.New()
	b.item.Name = name

	return b
}

// DefaultItems provides a set of default items.
func DefaultItems() []*grpctest.Item {
	return []*grpctest.Item{
		BuildItem().
			WithID(41).
			WithLocale("en-US").
			WithName("Item #41").
			New(),
		BuildItem().
			WithID(42).
			WithLocale("en-US").
			WithName("Item #42").
			New(),
	}
}

// DefaultItem provides an item for testing.
func DefaultItem() *grpctest.Item {
	return &grpctest.Item{
		Id:     41,
		Locale: "en-US",
		Name:   "Item #41",
	}
}

// BuildItem builds an item.
func BuildItem() ItemBuilder {
	return ItemBuilder{item: DefaultItem()}
}

func clone(src *grpctest.Item) *grpctest.Item {
	return proto.Clone(src).(*grpctest.Item)
}
