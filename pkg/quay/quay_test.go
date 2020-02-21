package quay

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
	data, _ := json.Marshal(payload)
	c.Request, err = http.NewRequest("POST", "/quay?secret=SECRET&account=cb1e73c5215b", bytes.NewBufferString(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	// setup mock
	hermesMock := new(HermesMock)
	eventURI := "registry:quay:namespace:name:push:cb1e73c5215b"
	event := hermes.NormalizedEvent{
		Original: string(data),
		Secret:   "SECRET",
		Variables: map[string]string{
			"namespace": "namespace",
			"name":      "name",
			"tag":       "updated_tags",
			"event":     "push",
			"url":       "homepage",
			"provider":  "quay",
			"type":      "registry",
		},
	}
	hermesMock.On("TriggerEvent", eventURI, &event).Return(nil)

	// bind quay to hermes API endpoint
	quay := NewQuay(hermesMock)
	router.POST("/quay", quay.HandleWebhook)
	router.HandleContext(c)

	// assert expectations
	hermesMock.AssertExpectations(t)
}
