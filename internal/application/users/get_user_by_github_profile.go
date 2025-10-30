package users

import "github.com/panoptescloud/api/internal/domain/users"

type GetUserByGithubNodeID struct {
	NodeID string
}

func (cmd GetUserByGithubNodeID) GetName() string {
	return "users.get_user_by_github_node_id"
}

func (uh *UserHandlers) GetUserByGithubNodeID(dto GetUserByGithubNodeID) (*users.User, error) {
	return uh.userRepo.ByGithubNodeId(dto.NodeID)
}
