package users

type SignInOrRegisterViaGithubDTO struct {
	Code string
}

type SignInOrRegisterViaGithubResponse struct {
	Token string
}

func (uh *UserHandlers) SigninOrRegisterViaGithub(dto SignInOrRegisterViaGithubDTO) (SignInOrRegisterViaGithubResponse, error) {
	token, err := uh.githubOauthClient.GetToken(dto.Code)

	if err != nil {
		return SignInOrRegisterViaGithubResponse{}, err
	}

	return SignInOrRegisterViaGithubResponse{
		Token: token,
	}, nil
}
