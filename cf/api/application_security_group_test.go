package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app security group api", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        ApplicationSecurityGroupRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewApplicationSecurityGroupRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".Create", func() {
		It("can create an app security group, given some attributes", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/app_security_groups",
				// FIXME: this matcher depend on the order of the key/value pairs in the map
				Matcher: testnet.RequestBodyMatcher(`{
					"name": "mygroup",
					"rules": [{"my-house": "my-rules"}],
					"space_guids": ["myspace"]
				}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})
			setupTestServer(req)

			err := repo.Create(models.ApplicationSecurityGroupFields{
				Name:       "mygroup",
				Rules:      []map[string]string{{"my-house": "my-rules"}},
				SpaceGuids: []string{"myspace"},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})

	Describe(".Read", func() {
		It("returns the app security group with the given name", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/app_security_groups?q=name:the-name&inline-relations-depth=1",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `
{
   "resources": [
      {
         "metadata": {
            "guid": "the-group-guid"
         },
         "entity": {
            "name": "the-name",
            "rules": [{"key": "value"}]
         }
      }
   ]
}
					`,
				},
			}))

			group, err := repo.Read("the-name")

			Expect(err).ToNot(HaveOccurred())
			Expect(group).To(Equal(models.ApplicationSecurityGroupFields{
				Name:  "the-name",
				Guid:  "the-group-guid",
				Rules: []map[string]string{{"key": "value"}},
			}))
		})

		It("returns a ModelNotFound error if the security group cannot be found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/app_security_groups?q=name:the-name&inline-relations-depth=1",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   `{"resources": []}`,
				},
			}))

			_, err := repo.Read("the-name")

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(errors.NewModelNotFoundError("model-type", "description")))
		})
	})

	Describe(".Delete", func() {
		It("deletes the application security group", func() {
			appSecurityGroupGuid := "the-security-group-guid"
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/app_security_groups/" + appSecurityGroupGuid,
				Response: testnet.TestResponse{
					Status: http.StatusNoContent,
				},
			}))

			err := repo.Delete(appSecurityGroupGuid)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})