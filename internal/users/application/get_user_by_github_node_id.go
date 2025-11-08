package users

import "github.com/panoptescloud/api/internal/users/domain"

type GetUserByGithubNodeID struct {
	NodeID string
}

func (cmd GetUserByGithubNodeID) GetName() string {
	return "users.query.get_user_by_github_node_id"
}

func (uh *UserQueryHandler) GetUserByGithubNodeID(dto GetUserByGithubNodeID) (*domain.User, error) {
	return uh.userRepo.ByGithubNodeId(dto.NodeID)
}
