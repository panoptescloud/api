package main_test

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/panoptescloud/api/internal/api/auth"
	"github.com/panoptescloud/api/internal/api/http"
	"github.com/panoptescloud/api/internal/api/http/v1beta"
	"github.com/panoptescloud/api/internal/common/bus"
	"github.com/panoptescloud/api/internal/infra/config"
	"github.com/panoptescloud/api/internal/infra/github_oauth"
	"github.com/panoptescloud/api/internal/infra/hasher"
	"github.com/panoptescloud/api/internal/infra/repository/postgres"
	"github.com/panoptescloud/api/tests/db/postgrestest"
	http_mocks "github.com/panoptescloud/api/tests/mocks/gen/api/http"
	"github.com/panoptescloud/api/tests/slogtest"
	testutil "github.com/panoptescloud/api/tests/util"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var sharedPool *pgxpool.Pool
var pgContainer *tcpostgres.PostgresContainer

type testapp struct {
	serverURL string
	dbPool    *pgxpool.Pool
	shutdown  func()
}

func TestMain(m *testing.M) {
	pgContainer, sharedPool = postgrestest.SetupTestDB()

	exitCode := m.Run()

	// defer doesn't work properly because of the os.Exit call, need to do it manually
	sharedPool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	if err := pgContainer.Terminate(ctx); err != nil {
		log.Errorf("failed to terminate container: %v", err)
	}
	cancel()

	os.Exit(exitCode)
}

func setupServer(t *testing.T, appCfg *config.Config) testapp {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("recovered from panic")
			t.FailNow()
		}
	}()

	pool := postgrestest.PrepareDBForTest(t, sharedPool, true)

	sm := http_mocks.NewMockSessionManager(t)
	al := http_mocks.NewMockActorLoader(t)
	l, _ := slogtest.NewLogger()

	srv := http.NewServer(sm, al, l, l)

	b := bus.New()
	ghOauthClient := github_oauth.NewClient(
		appCfg.GetGithubOauthClientId(),
		appCfg.GetGithubOauthClientSecret(),
		l.With("component", "github-oauth-client"),
	)

	refreshTokensRepo := postgres.NewRefreshTokensRepository(pool)
	h := hasher.NewHMACSHA256Hasher([]byte(appCfg.GetAuthHMACKey()))
	sessionManager, err := auth.NewSessionManager(
		refreshTokensRepo,
		h,
		appCfg.GetAuthJWTPrivateKeyPath(),
		appCfg.GetAuthJWTPublicKeyPath(),
	)

	require.Nil(t, err)

	srv.Initialise([]http.Controller{
		v1beta.NewProbesController(),
		v1beta.NewAuthController(
			b,
			ghOauthClient,
			sessionManager,
			l,
		),
	})

	e := srv.GetEcho()

	testSrv := httptest.NewServer(e)

	return testapp{
		serverURL: testSrv.URL,
		dbPool:    pool,
		shutdown: func() {
			testSrv.Close()
			pool.Close()
		},
	}
}

func defaultConfig() *config.Config {
	appRoot := testutil.GetAppRoot()
	cfg := config.Default()
	cfg.Auth.HMACKey = "GS/6rngzu5KUr8bqYJELEqrzD/qeaxC2PBljE78+3GM="
	cfg.Auth.JWT.PrivateKeyFile = fmt.Sprintf("%s/tests/fixtures/keys/private.key", appRoot)
	cfg.Auth.JWT.PublicKeyFile = fmt.Sprintf("%s/tests/fixtures/keys/public.key", appRoot)

	return cfg
}

func ensureIsStarted(t *testing.T, app testapp) {
	resp, err := stdhttp.Get(app.serverURL + "/api/v1beta/_probes/startup")

	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, stdhttp.StatusNoContent, resp.StatusCode)
}

func Test_Api_V1Beta_StartupProbe(t *testing.T) {
	t.Parallel()

	app := setupServer(t, defaultConfig())
	defer app.shutdown()

	// This handles the testing here
	ensureIsStarted(t, app)
}
