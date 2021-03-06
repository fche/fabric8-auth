package controller_test

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-auth/account"
	"github.com/fabric8-services/fabric8-auth/app"
	"github.com/fabric8-services/fabric8-auth/app/test"
	"github.com/fabric8-services/fabric8-auth/application"
	. "github.com/fabric8-services/fabric8-auth/controller"
	"github.com/fabric8-services/fabric8-auth/gormapplication"
	"github.com/fabric8-services/fabric8-auth/gormtestsupport"
	"github.com/fabric8-services/fabric8-auth/login"
	"github.com/fabric8-services/fabric8-auth/resource"
	testsupport "github.com/fabric8-services/fabric8-auth/test"
	testtoken "github.com/fabric8-services/fabric8-auth/test/token"
	"github.com/fabric8-services/fabric8-auth/token/oauth"

	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestLoginREST struct {
	gormtestsupport.DBTestSuite

	db *gormapplication.GormDB
}

func TestRunLoginREST(t *testing.T) {
	suite.Run(t, &TestLoginREST{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

func (rest *TestLoginREST) SetupTest() {
	rest.DBTestSuite.SetupTest()
	rest.db = gormapplication.NewGormDB(rest.DB)
}

func (rest *TestLoginREST) UnSecuredController() (*goa.Service, *LoginController) {
	svc := testsupport.ServiceAsUser("Login-Service", testsupport.TestIdentity)
	return svc, &LoginController{Controller: svc.NewController("login"), Auth: TestLoginService{}, Configuration: rest.Configuration}
}

func (rest *TestLoginREST) SecuredController() (*goa.Service, *LoginController) {
	loginService := newTestKeycloakOAuthProvider(rest.db, rest.Configuration)

	svc := testsupport.ServiceAsUser("Login-Service", testsupport.TestIdentity)
	return svc, NewLoginController(svc, loginService, loginService.TokenManager, rest.Configuration)
}

func newTestKeycloakOAuthProvider(db application.DB, configuration LoginConfiguration) *login.KeycloakOAuthProvider {
	return login.NewKeycloakOAuthProvider(db.Identities(), db.Users(), testtoken.TokenManager, db)
}

func (rest *TestLoginREST) TestLoginOK() {
	t := rest.T()
	resource.Require(t, resource.UnitTest)
	svc, ctrl := rest.UnSecuredController()

	test.LoginLoginTemporaryRedirect(t, svc.Context, svc, ctrl, nil, nil, nil, nil)
}

func (rest *TestLoginREST) TestOfflineAccessOK() {
	t := rest.T()
	resource.Require(t, resource.UnitTest)
	svc, ctrl := rest.UnSecuredController()

	offline := "offline_access"
	resp := test.LoginLoginTemporaryRedirect(t, svc.Context, svc, ctrl, nil, nil, nil, &offline)
	assert.Contains(t, resp.Header().Get("Location"), "scope=offline_access")

	resp = test.LoginLoginTemporaryRedirect(t, svc.Context, svc, ctrl, nil, nil, nil, nil)
	assert.NotContains(t, resp.Header().Get("Location"), "scope=offline_access")
}

type TestLoginService struct{}

func (t TestLoginService) Perform(ctx *app.LoginLoginContext, config oauth.OauthConfig, serviceConfig login.LoginServiceConfiguration) error {
	return ctx.TemporaryRedirect()
}

func (t TestLoginService) CreateOrUpdateIdentity(ctx context.Context, accessToken string) (*account.Identity, bool, error) {
	return nil, false, nil
}

func (t TestLoginService) Link(ctx *app.LinkLinkContext, brokerEndpoint string, clientID string, validRedirectURL string) error {
	return ctx.TemporaryRedirect()
}

func (t TestLoginService) LinkSession(ctx *app.SessionLinkContext, brokerEndpoint string, clientID string, validRedirectURL string) error {
	return ctx.TemporaryRedirect()
}

func (t TestLoginService) LinkCallback(ctx *app.CallbackLinkContext, brokerEndpoint string, clientID string) error {
	return ctx.TemporaryRedirect()
}
