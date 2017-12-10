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
		TriggerEvent(event *NormalizedEvent) error
	}

	// APIEndpoint Hermes API endpoint
	APIEndpoint struct {
		endpoint *sling.Sling
	}

	// NormalizedEvent normalized event: {event-uri, original-payload, secret, variables-map}
	NormalizedEvent struct {
		EventURI  string            `json:"event"`
		Original  string            `json:"original,omitempty"`
		Secret    string            `json:"secret,omitempty"`
		Variables map[string]string `json:"variables,omitempty"`
	}
)

// NewHermesEndpoint create new Hermes API endpoint from url and API token
func NewHermesEndpoint(url, token string) Service {
	endpoint := sling.New().Base(url).Set("x-access-token", token)
	return &APIEndpoint{endpoint}
}

// TriggerEvent send normalized event to Hermes trigger-manager server
func (api *APIEndpoint) TriggerEvent(event *NormalizedEvent) error {
	log.Debugf("Triggering event '%s'", event.EventURI)
	resp, err := api.endpoint.New().Post(fmt.Sprint("trigger/", event.EventURI)).BodyJSON(event).ReceiveSuccess(nil)
	if err != nil {
		log.Error(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: error triggering event '%s'", resp.Status, event.EventURI)
	}
	log.Debugf("Event '%s' triggered", event.EventURI)
	return nil
}
