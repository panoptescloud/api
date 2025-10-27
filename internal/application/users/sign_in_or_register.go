package users

type SignInOrRegisterViaGithub struct {
	Code string
}

func (cmd SignInOrRegisterViaGithub) GetName() string {
	return "users.sign_in_or_register.github"
}

type SignInOrRegisterViaGithubResponse struct {
	Token string
}

func (uh *UserHandlers) SigninOrRegisterViaGithub(dto SignInOrRegisterViaGithub) error {
	_, err := uh.githubOauthClient.GetToken(dto.Code)

	return err
}
