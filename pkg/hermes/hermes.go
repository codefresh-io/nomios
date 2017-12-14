package hermes

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	log "github.com/sirupsen/logrus"
)

type (
	// Service Codefresh Service
	Service interface {
		TriggerEvent(eventURI string, event *NormalizedEvent) error
	}

	// APIEndpoint Hermes API endpoint
	APIEndpoint struct {
		endpoint *sling.Sling
	}

	// NormalizedEvent normalized event: {event-uri, original-payload, secret, variables-map}
	NormalizedEvent struct {
		Original  string            `json:"original,omitempty"`
		Secret    string            `json:"secret,omitempty"`
		Variables map[string]string `json:"variables,omitempty"`
	}

	ServiceError struct {
		Status int    `json:"status"`
		Title  string `json:"message"`
		Body   string `json:"error"`
	}
)

// NewNormalizedEvent init NormalizedEvent struct
func NewNormalizedEvent() *NormalizedEvent {
	var event NormalizedEvent
	event.Variables = make(map[string]string)
	return &event
}

// NewHermesEndpoint create new Hermes API endpoint from url and API token
func NewHermesEndpoint(url, token string) Service {
	endpoint := sling.New().Base(url).Set("x-access-token", token)
	return &APIEndpoint{endpoint}
}

// TriggerEvent send normalized event to Hermes trigger-manager server
func (api *APIEndpoint) TriggerEvent(eventURI string, event *NormalizedEvent) error {
	log.Debugf("Triggering event '%s'", eventURI)
	var runs []string
	var failure ServiceError
	resp, err := api.endpoint.New().Post(fmt.Sprint("trigger/", eventURI)).BodyJSON(event).Receive(&runs, &failure)
	if err != nil {
		log.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: error triggering event '%s'", resp.Status, eventURI)
	}
	log.Debugf("Event '%s' triggered", eventURI)
	log.Debugf("Running pipelines: %v", runs)
	return nil
}
