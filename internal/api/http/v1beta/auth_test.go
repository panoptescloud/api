package v1beta_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	v1beta_mocks "github.com/panoptescloud/api/tests/mocks/gen/api/http/v1beta"
	"github.com/panoptescloud/api/tests/slogtest"
	testutil "github.com/panoptescloud/api/tests/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/panoptescloud/api/internal/api/http/v1beta"
	"github.com/panoptescloud/api/internal/common"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/common/dto"
	"github.com/panoptescloud/api/internal/common/validation"
	usersapp "github.com/panoptescloud/api/internal/users/application"
	"github.com/panoptescloud/api/internal/users/domain"
)

func Test_AuthController_LoginWithGithub(t *testing.T) {
	userUid := testutil.NewUuidV7(t)
	issuedAt := time.Date(2025, time.November, 8, 12, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2025, time.November, 8, 12, 5, 0, 0, time.UTC)
	refreshExpiresAt := time.Date(2025, time.November, 15, 12, 0, 0, 0, time.UTC)
	refreshTokenUid := testutil.NewUuidV7(t)

	tests := []struct {
		name                  string
		code                  string
		prepareGithubClient   func(*testing.T, *v1beta_mocks.MockGithubOauthClient)
		prepareSessionManager func(*testing.T, *v1beta_mocks.MockSessionManager)
		prepareBus            func(*testing.T, *bus.Bus)
		assertLogs            func(*testing.T, *slogtest.MockHandler)
		expect                *v1beta.SessionResponse
		expectErr             error
	}{
		{
			name: "error getting token",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", errors.New("failed to get token"))
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus:            func(tt *testing.T, b *bus.Bus) {},
			expect:                nil,
			expectErr:             errors.New("failed to get token"),
		},
		{
			name: "error getting profile",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(
					domain.GithubProfile{},
					errors.New("error getting profile"),
				)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus:            func(tt *testing.T, b *bus.Bus) {},
			expect:                nil,
			expectErr:             errors.New("error getting profile"),
		},
		{
			name: "error getting user",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					require.Equal(t, "some_node_id", dto.NodeID, "node id passed to GetUserByGithubNodeID was incorrect")
					return nil, errors.New("error getting user")
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
			},
			expect:    nil,
			expectErr: errors.New("error getting user"),
		},
		{
			name: "error creating user",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					return nil, nil
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return errors.New("failed to create user")
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			expect:    nil,
			expectErr: errors.New("failed to create user"),
		},
		{
			// Any violation with an empty violation rught now returns a 500. This
			// is because something went wron in our side, this info should come
			// from gitub. If they're empty it means our integration with github
			// isn't as solid as we thought. There's not really anything a client
			// can do at this stage, so it's not a client level error where they
			// can update the payload or anything.
			name: "error creating user/validation error/random empty field",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					return nil, nil
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return validation.Error{
						FieldErrors: []validation.FieldError{
							{
								Key: "Random",
								Errors: validation.Violations{
									validation.NotEmptyViolation{},
								},
							},
						},
					}
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			assertLogs: func(tt *testing.T, h *slogtest.MockHandler) {
				h.MustHaveSeen(tt, "empty fields during user creation", slog.LevelError, map[string]any{
					"field_errors": []validation.FieldError{
						{
							Key: "Random",
							Errors: validation.Violations{
								validation.NotEmptyViolation{},
							},
						},
					},
				})
			},
			expect:    nil,
			expectErr: common.ErrInternalError{},
		},
		{
			// Similar to the above, if the email is invalid it means that our
			// email validation differs to what github considers a valid email.
			// Nothing a client can do, and we'll need to fix/sync validation
			// to support everything github does.
			name: "error creating user/validation error/invalid email",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					return nil, nil
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return validation.Error{
						FieldErrors: []validation.FieldError{
							{
								Key: "Email",
								Errors: validation.Violations{
									validation.EmailViolation{},
								},
							},
						},
					}
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			assertLogs: func(tt *testing.T, h *slogtest.MockHandler) {
				h.MustHaveSeen(tt, "invalid email during user creation", slog.LevelError, map[string]any{
					"field_errors": []validation.FieldError{
						{
							Key: "Email",
							Errors: validation.Violations{
								validation.EmailViolation{},
							},
						},
					},
				})
			},
			expect:    nil,
			expectErr: common.ErrInternalError{},
		},
		{
			// The two above warrant specific tests, as they need to be handled
			// this way. There will be other cases where we need to return something
			// like a 409 (if email is already used). But right now they're not
			// implemented, so any validation error right now just gives an
			// internal error.
			name: "error creating user/validation error/catch all",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					return nil, nil
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return validation.Error{}
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			assertLogs: func(tt *testing.T, h *slogtest.MockHandler) {
				h.MustHaveSeen(tt, "error during user creation", slog.LevelError, map[string]any{
					"err": validation.Error{},
				})
			},
			expect:    nil,
			expectErr: common.ErrInternalError{},
		},
		{
			name: "fails to retrieve user after successful creation",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				calls := 0
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					calls++
					require.Equal(t, "some_node_id", dto.NodeID, "node id passed to GetUserByGithubNodeID was incorrect")
					switch calls {
					case 1:
						return nil, nil
					case 2:
						return nil, errors.New("failed to get user, second time")
					default:
						tt.Log("Unexpected query GetUserByGithubNodeID, called too many times")
						tt.FailNow()
						return nil, nil
					}
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return nil
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			expect:    nil,
			expectErr: errors.New("failed to get user, second time"),
		},
		{
			name: "fails to create session",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().Create(mock.MatchedBy(func(arg domain.UserID) bool {
					return arg.String() == userUid.String()
				})).Return(dto.Session{}, errors.New("failed to create session"))
			},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				calls := 0
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					calls++
					require.Equal(t, "some_node_id", dto.NodeID, "node id passed to GetUserByGithubNodeID was incorrect")
					switch calls {
					case 1:
						return nil, nil
					case 2:
						return domain.HydrateUser(
							userUid.String(),
							"James Holden",
							"james.holden@org.unn",
							"some_node_id",
						)
					default:
						t.Log("Unexpected query GetUserByGithubNodeID, called too many times")
						t.FailNow()
						return nil, nil
					}
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return nil
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			expect:    nil,
			expectErr: errors.New("failed to create session"),
		},
		{
			name: "creates new user",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().Create(mock.MatchedBy(func(arg domain.UserID) bool {
					return arg.String() == userUid.String()
				})).Return(dto.Session{
					JWT: "somejwt",
					RefreshToken: dto.RefreshToken{
						ID: refreshTokenUid,
						Token: dto.HashedValue{
							// note, while it would be available, we wanna ensure the Original
							// is not used
							Value: "somerefresh",
						},
						IssuedAt:  issuedAt,
						ExpiresAt: refreshExpiresAt,
						UserID:    userUid,
					},
					CSRFToken:    "somecsrf",
					UserID:       userUid.String(),
					IssuedAt:     issuedAt,
					JWTExpiresAt: expiresAt,
				}, nil)
			},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				calls := 0
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					calls++
					require.Equal(t, "some_node_id", dto.NodeID, "node id passed to GetUserByGithubNodeID was incorrect")
					switch calls {
					case 1:
						return nil, nil
					case 2:
						return domain.HydrateUser(
							userUid.String(),
							"James Holden",
							"james.holden@org.unn",
							"some_node_id",
						)
					default:
						t.Log("Unexpected query GetUserByGithubNodeID, called too many times")
						t.FailNow()
						return nil, nil
					}
				}

				dummyCreateUserHandler := func(dto usersapp.CreateUser) error {
					return nil
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
				bus.RegisterCommand(b, dummyCreateUserHandler)
			},
			expect: &v1beta.SessionResponse{
				SetCookie: []http.Cookie{
					{
						Name:        "auth_token",
						Value:       "somejwt",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    true,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "csrf_token",
						Value:       "somecsrf",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "refresh_token",
						Value:       "somerefresh",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      604800,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
				},
				Body: v1beta.SessionResponseBody{
					Data: v1beta.SessionResponseData{
						CSRFToken: "somecsrf",
						User: v1beta.User{
							ID: userUid.String(),
						},
					},
				},
			},
			expectErr: nil,
		},
		{
			name: "user already exists",
			code: "blah",
			prepareGithubClient: func(tt *testing.T, gh *v1beta_mocks.MockGithubOauthClient) {
				gh.EXPECT().GetToken("blah").Return("sometoken", nil)
				gh.EXPECT().GetProfile("sometoken").Return(domain.GithubProfile{
					NodeID: "some_node_id",
					Email:  "james.holden@org.unn",
					Name:   "JamesHolden",
				}, nil)
			},
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().Create(mock.MatchedBy(func(arg domain.UserID) bool {
					return arg.String() == userUid.String()
				})).Return(dto.Session{
					JWT: "somejwt",
					RefreshToken: dto.RefreshToken{
						ID: refreshTokenUid,
						Token: dto.HashedValue{
							// note, while it would be available, we wanna ensure the Original
							// is not used
							Value: "somerefresh",
						},
						IssuedAt:  issuedAt,
						ExpiresAt: refreshExpiresAt,
						UserID:    userUid,
					},
					CSRFToken:    "somecsrf",
					UserID:       userUid.String(),
					IssuedAt:     issuedAt,
					JWTExpiresAt: expiresAt,
				}, nil)
			},
			prepareBus: func(tt *testing.T, b *bus.Bus) {
				dummyGetUserByGithubNodeIDHandler := func(dto usersapp.GetUserByGithubNodeID) (*domain.User, error) {
					return domain.HydrateUser(
						userUid.String(),
						"James Holden",
						"james.holden@org.unn",
						"some_node_id",
					)
				}

				bus.RegisterQuery(b, dummyGetUserByGithubNodeIDHandler)
			},
			expect: &v1beta.SessionResponse{
				SetCookie: []http.Cookie{
					{
						Name:        "auth_token",
						Value:       "somejwt",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    true,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "csrf_token",
						Value:       "somecsrf",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "refresh_token",
						Value:       "somerefresh",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      604800,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
				},
				Body: v1beta.SessionResponseBody{
					Data: v1beta.SessionResponseData{
						CSRFToken: "somecsrf",
						User: v1beta.User{
							ID: userUid.String(),
						},
					},
				},
			},
			expectErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			b := bus.New()
			gh := v1beta_mocks.NewMockGithubOauthClient(tt)
			sm := v1beta_mocks.NewMockSessionManager(tt)
			l, lh := slogtest.NewLogger()

			test.prepareGithubClient(tt, gh)
			test.prepareSessionManager(tt, sm)
			test.prepareBus(tt, b)

			c := v1beta.NewAuthController(b, gh, sm, l)
			resp, err := c.LoginWithGithub(context.TODO(), &v1beta.GithubLoginRequest{
				Body: v1beta.GithubLoginRequestBody{
					Code: test.code,
				},
			})

			require.Equal(tt, test.expectErr, err)
			assert.Equal(tt, test.expect, resp)

			if test.assertLogs != nil {
				test.assertLogs(tt, lh)
			}
		})
	}
}

func Test_AuthController_Logout(t *testing.T) {
	tests := []struct {
		name                  string
		refreshToken          string
		expect                *v1beta.LogoutResponse
		expectErr             error
		prepareSessionManager func(*testing.T, *v1beta_mocks.MockSessionManager)
	}{
		{
			name:                  "missing refresh token",
			refreshToken:          "",
			prepareSessionManager: func(*testing.T, *v1beta_mocks.MockSessionManager) {},
			expectErr:             common.ErrUnauthorised{},
			expect:                nil,
		},
		{
			name:         "error deleting refresh token",
			refreshToken: "blah",
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().DeleteRefreshToken(dto.HashedValue{
					Value: "blah",
				}).Return(errors.New("failed to delete refresh token"))
			},
			expectErr: errors.New("failed to delete refresh token"),
			expect: &v1beta.LogoutResponse{
				Status: 204,
				SetCookie: []http.Cookie{
					{
						Name:     "auth_token",
						Value:    "",
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
					{
						Name:     "csrf_token",
						Value:    "",
						Path:     "/",
						HttpOnly: false,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
					{
						Name:     "refresh_token",
						Value:    "",
						Path:     "/",
						HttpOnly: false,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
				},
			},
		},
		{
			name:         "successful logout",
			refreshToken: "blah",
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().DeleteRefreshToken(dto.HashedValue{
					Value: "blah",
				}).Return(nil)
			},
			expectErr: nil,
			expect: &v1beta.LogoutResponse{
				Status: 204,
				SetCookie: []http.Cookie{
					{
						Name:     "auth_token",
						Value:    "",
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
					{
						Name:     "csrf_token",
						Value:    "",
						Path:     "/",
						HttpOnly: false,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
					{
						Name:     "refresh_token",
						Value:    "",
						Path:     "/",
						HttpOnly: false,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   -1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			b := bus.New()
			gh := v1beta_mocks.NewMockGithubOauthClient(tt)
			l, _ := slogtest.NewLogger()
			sm := v1beta_mocks.NewMockSessionManager(tt)
			test.prepareSessionManager(tt, sm)

			c := v1beta.NewAuthController(b, gh, sm, l)

			resp, err := c.Logout(context.TODO(), &v1beta.LogoutRequest{
				RefreshToken: test.refreshToken,
			})

			require.Equal(tt, test.expectErr, err)
			assert.Equal(tt, test.expect, resp)
		})
	}
}

func Test_AuthController_Refresh(t *testing.T) {
	userUid := testutil.NewUuidV7(t)
	issuedAt := time.Date(2025, time.November, 8, 12, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2025, time.November, 8, 12, 5, 0, 0, time.UTC)
	refreshExpiresAt := time.Date(2025, time.November, 15, 12, 0, 0, 0, time.UTC)
	refreshTokenUid := testutil.NewUuidV7(t)

	tests := []struct {
		name                  string
		refreshToken          string
		expect                *v1beta.SessionResponse
		expectErr             error
		prepareSessionManager func(*testing.T, *v1beta_mocks.MockSessionManager)
	}{
		{
			name:                  "missing refresh token",
			refreshToken:          "",
			prepareSessionManager: func(*testing.T, *v1beta_mocks.MockSessionManager) {},
			expectErr:             common.ErrUnauthorised{},
			expect:                nil,
		},
		{
			name:         "error refreshing session",
			refreshToken: "blah",
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().Refresh(dto.HashedValue{
					Value: "blah",
				}).Return(dto.Session{}, errors.New("failed to refresh token"))
			},
			expectErr: errors.New("failed to refresh token"),
			expect:    nil,
		},
		{
			name:         "successful refresh",
			refreshToken: "blah",
			prepareSessionManager: func(tt *testing.T, sm *v1beta_mocks.MockSessionManager) {
				sm.EXPECT().Refresh(dto.HashedValue{
					Value: "blah",
				}).Return(dto.Session{
					JWT: "somejwt",
					RefreshToken: dto.RefreshToken{
						ID: refreshTokenUid,
						Token: dto.HashedValue{
							// note, while it would be available, we wanna ensure the Original
							// is not used
							Value: "somerefresh",
						},
						IssuedAt:  issuedAt,
						ExpiresAt: refreshExpiresAt,
						UserID:    userUid,
					},
					CSRFToken:    "somecsrf",
					UserID:       userUid.String(),
					IssuedAt:     issuedAt,
					JWTExpiresAt: expiresAt,
				}, nil)
			},
			expectErr: nil,
			expect: &v1beta.SessionResponse{
				SetCookie: []http.Cookie{
					{
						Name:        "auth_token",
						Value:       "somejwt",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    true,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "csrf_token",
						Value:       "somecsrf",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      300,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
					{
						Name:        "refresh_token",
						Value:       "somerefresh",
						Quoted:      false,
						Path:        "/",
						Domain:      "",
						Expires:     time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
						RawExpires:  "",
						MaxAge:      604800,
						Secure:      true,
						HttpOnly:    false,
						SameSite:    http.SameSiteStrictMode,
						Partitioned: false,
						Raw:         "",
						Unparsed:    nil,
					},
				},
				Body: v1beta.SessionResponseBody{
					Data: v1beta.SessionResponseData{
						CSRFToken: "somecsrf",
						User: v1beta.User{
							ID: userUid.String(),
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			b := bus.New()
			gh := v1beta_mocks.NewMockGithubOauthClient(tt)
			l, _ := slogtest.NewLogger()
			sm := v1beta_mocks.NewMockSessionManager(tt)
			test.prepareSessionManager(tt, sm)

			c := v1beta.NewAuthController(b, gh, sm, l)

			resp, err := c.Refresh(context.TODO(), &v1beta.RefreshRequest{
				RefreshToken: test.refreshToken,
			})

			require.Equal(tt, test.expectErr, err)
			assert.Equal(tt, test.expect, resp)
		})
	}
}
