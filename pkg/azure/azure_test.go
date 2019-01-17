package azure

import (
	"bytes"
	"encoding/json"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HermesMock struct {
	mock.Mock
}

func (m *HermesMock) TriggerEvent(eventURI string, event *hermes.NormalizedEvent) error {
	args := m.Called(eventURI, event)
	return args.Error(0)

}

func TestContextBindWithQuery(t *testing.T) {
	rr := httptest.NewRecorder()
	c, router := gin.CreateTestContext(rr)

	file, err := ioutil.ReadFile("./test_payload.json")
	if err != nil {
		t.Fatal(err)
	}

	var payload webhookPayload
	err = json.Unmarshal(file, &payload)
	if err != nil {
		t.Fatal(err)
	}
	payload.Target.Name = "repo"
	data, _ := json.Marshal(payload)
	c.Request, err = http.NewRequest("POST", "/azure?secret=SECRET&account=cb1e73c5215b", bytes.NewBufferString(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	// setup mock
	hermesMock := new(HermesMock)
	eventURI := "registry:azure:host:repo:push:cb1e73c5215b"
	event := hermes.NormalizedEvent{
		Original: string(data),
		Secret:   "SECRET",
		Variables: map[string]string{
			"event":     "push",
			"namespace": "host",
			"name":      "repo",
			"tag":       "latest",
			"pushed_at": "2018-11-05T18:24:27.609016022Z",
		},
	}
	hermesMock.On("TriggerEvent", eventURI, &event).Return(nil)

	// bind dockerhub to hermes API endpoint
	azure := NewAzure(hermesMock)
	router.POST("/azure", azure.HandleWebhook)
	router.HandleContext(c)

	// assert expectations
	hermesMock.AssertExpectations(t)
}
