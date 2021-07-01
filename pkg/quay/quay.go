package quay

import (
	"encoding/json"
	"fmt"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Quay struct {
	hermesSvc hermes.Service
}

func NewQuay(svc hermes.Service) *Quay {
	return &Quay{svc}
}

type webhookPayload struct {
	Name             string   `json:"name"`
	Repository       string   `json:"repository"`
	DockerURL        string   `json:"docker_url"`
	Namespace        string   `json:"namespace"`
	PrunedImageCount int64    `json:"pruned_image_count"`
	Homepage         string   `json:"homepage"`
	UpdatedTags      []string `json:"updated_tags"`
}

func constructEventURI(payload *webhookPayload, account string) string {
	uri := fmt.Sprintf("registry:quay:%s:%s:push", payload.Namespace, payload.Name)
	if account != "" {
		uri = fmt.Sprintf("%s:%s", uri, account)
	}
	return uri
}

func (q *Quay) HandleWebhook(c *gin.Context) {
	payload := webhookPayload{}
	if err := c.BindJSON(&payload); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Name: " + payload.Name)
	fmt.Println("Namespace: " + payload.Namespace)

	event := hermes.NewNormalizedEvent()
	eventURI := constructEventURI(&payload, c.Query("account"))
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Error("Failed to covert webhook payload structure to JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// keep original JSON
	event.Original = string(payloadJSON)

	// get image push details
	event.Variables["namespace"] = payload.Namespace
	event.Variables["name"] = payload.Name
	//TODO : handle array of tags
	if payload.UpdatedTags != nil && len(payload.UpdatedTags) > 0 {
		event.Variables["tag"] = payload.UpdatedTags[0]
	}
	event.Variables["event"] = "push"
	event.Variables["url"] = payload.Homepage
	event.Variables["provider"] = "quay"
	event.Variables["type"] = "registry"

	// get secret from URL query
	event.Secret = c.Query("secret")

	log.Debug("Event url " + eventURI)

	// invoke trigger
	err = q.hermesSvc.TriggerEvent(eventURI, event)
	if err != nil {
		log.WithError(err).Error("Failed to trigger event pipelines")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
