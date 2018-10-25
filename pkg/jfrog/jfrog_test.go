package jfrog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
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
	c.Request, err = http.NewRequest("POST", "/jfrog?secret=SECRET&account=cb1e73c5215b", bytes.NewBufferString(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	// setup mock
	hermesMock := new(HermesMock)
	eventURI := "registry:jfrog:local:test:push:cb1e73c5215b"
	event := hermes.NormalizedEvent{
		Original: string(data),
		Secret:   "SECRET",
		Variables: map[string]string{
			"event":     "docker.tagCreated",
			"namespace": "local",
			"name":      "test",
			"tag":       "tagName",
			"pusher":    "admin",
			"pushed_at": time.Unix(1540479021, 0).Format(time.RFC3339),
		},
	}
	hermesMock.On("TriggerEvent", eventURI, &event).Return(nil)

	// bind dockerhub to hermes API endpoint
	jfrog := NewJFrog(hermesMock)
	router.POST("/jfrog", jfrog.HandleWebhook)
	router.HandleContext(c)

	// assert expectations
	hermesMock.AssertExpectations(t)
}
