package users

import (
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/organisations/domain"
	users "github.com/panoptescloud/api/internal/users/application"
	usersdomain "github.com/panoptescloud/api/internal/users/domain"
)

// Ideallistically, this shouldn't be using the bus. The main reason is that this
// means that the current context (organisations) has to reach into the users
// context to know the dto that should be used to run the query. In reality this
// should probably be going through something like grpc, but honestly I can't
// bring myself to add that much drama just to satisfy idealism in the design
// right now.
//
// At least two better options:
//  1. Use async events to keep a "view" of the users in the organisations
//     context. Probably preferred as it means less complexity in the deployments,
//     and we'll need to listen to events from users context eventually. Does
//     introduce some level of eventual consistency though.
//  2. Use gRPC/HTTP etc, and pull the data from another service. Don't like this
//     mostly because it means adding a whole lot of orchestration complexity as
//     well as network latency for a simple lookup. But if the contexts were deployed
//     separately already this would be a pretty valid path.
type UsersBridge struct {
	bus *bus.Bus
}

func (b UsersBridge) UserExists(id domain.MemberID) (bool, error) {
	userID, err := usersdomain.NewUserID(id.String())

	if err != nil {
		return false, err
	}

	user, err := bus.RunQuery[users.GetUserByID, *usersdomain.User](b.bus, users.GetUserByID{
		ID: userID,
	})

	return user != nil, err
}

func NewUsersBridge(b *bus.Bus) *UsersBridge {
	return &UsersBridge{
		bus: b,
	}
}
