package resource_test

import (
	"testing"

	"github.com/fabric8-services/fabric8-auth/account"
	"github.com/fabric8-services/fabric8-auth/authorization/resource"
	"github.com/fabric8-services/fabric8-auth/errors"
	"github.com/fabric8-services/fabric8-auth/gormtestsupport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/satori/go.uuid"

	res "github.com/fabric8-services/fabric8-auth/resource"
)

type resourceBlackBoxTest struct {
	gormtestsupport.DBTestSuite
	repo             resource.ResourceRepository
	identityRepo     account.IdentityRepository
	resourceTypeRepo resource.ResourceTypeRepository
}

func TestRunResourceBlackBoxTest(t *testing.T) {
	suite.Run(t, &resourceTypeBlackBoxTest{DBTestSuite: gormtestsupport.NewDBTestSuite("../../config.yaml")})
}

func (s *resourceBlackBoxTest) SetupTest() {
	s.DBTestSuite.SetupTest()
	s.repo = resource.NewResourceRepository(s.DB)
	s.identityRepo = account.NewIdentityRepository(s.DB)
	s.resourceTypeRepo = resource.NewResourceTypeRepository(s.DB)
}

func (s *resourceBlackBoxTest) TestOKToDelete() {
	t := s.T()
	res.Require(t, res.Database)

	resource := createAndLoadResource(s)
	createAndLoadResource(s)

	loadedResource, err := s.repo.Load(s.Ctx, resource.ResourceID)
	require.NotNil(s.T(), loadedResource, "Created resource should be loaded")
	require.Nil(s.T(), err, "Should be no error when loading existing resource")

	err = s.repo.Delete(s.Ctx, resource.ResourceID)
	assert.Nil(s.T(), err, "Should be no error when deleting resource")

	// Check the resource is deleted correctly
	loadedResource, err = s.repo.Load(s.Ctx, resource.ResourceID)
	require.Nil(s.T(), loadedResource, "Deleted resource should not be possible to load")
	require.NotNil(s.T(), err, "Should be error when loading non-existing resource")
}

func (s *resourceBlackBoxTest) TestOKToLoad() {
	t := s.T()
	res.Require(t, res.Database)

	createAndLoadResource(s)
}

func (s *resourceBlackBoxTest) TestExistsResource() {
	t := s.T()
	res.Require(t, res.Database)

	t.Run("resource exists", func(t *testing.T) {
		//t.Parallel()
		resource := createAndLoadResource(s)
		// when
		err := s.repo.CheckExists(s.Ctx, resource.ResourceID)
		// then
		require.Nil(t, err)
	})

	t.Run("resource doesn't exist", func(t *testing.T) {
		//t.Parallel()
		// Check not existing
		err := s.repo.CheckExists(s.Ctx, uuid.NewV4().String())
		// then
		require.IsType(s.T(), errors.NotFoundError{}, err)
	})
}

func (s *resourceBlackBoxTest) TestOKToSave() {
	t := s.T()
	res.Require(t, res.Database)

	resource := createAndLoadResource(s)

	resource.Description = "foo"
	err := s.repo.Save(s.Ctx, resource)
	require.Nil(s.T(), err, "Could not update resource")

	updatedResource, err := s.repo.Load(s.Ctx, resource.ResourceID)
	require.Nil(s.T(), err, "Could not load resource")
	assert.Equal(s.T(), updatedResource.Description, "foo")
}

func createAndLoadResource(s *resourceBlackBoxTest) *resource.Resource {
	identity := &account.Identity{
		ID:           uuid.NewV4(),
		Username:     "resource_blackbox_test_someuserTestIdentity2",
		ProviderType: account.KeycloakIDP}

	err := s.identityRepo.Create(s.Ctx, identity)
	require.Nil(s.T(), err, "Could not create identity")

	resourceType := &resource.ResourceType{
		ResourceTypeID: uuid.NewV4(),
		Name:           "resource_blackbox_test_Area" + uuid.NewV4().String(),
		Description:    "An area is a logical grouping within a space",
	}

	err = s.resourceTypeRepo.Create(s.Ctx, resourceType)
	require.Nil(s.T(), err, "Could not create resource type")

	resource := &resource.Resource{
		ResourceID:     uuid.NewV4().String(),
		ParentResource: nil,
		Owner:          *identity,
		ResourceType:   *resourceType,
		Description:    "resource_blackbox_test_A description of the created resource",
	}

	err = s.repo.Create(s.Ctx, resource)
	require.Nil(s.T(), err, "Could not create resource")

	createdResource, err := s.repo.Load(s.Ctx, resource.ResourceID)
	require.Nil(s.T(), err, "Could not load resource")

	require.Equal(s.T(), resource.Description, createdResource.Description)

	return createdResource
}
