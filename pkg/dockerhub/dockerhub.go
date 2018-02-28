package dockerhub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// DockerHub struct
type DockerHub struct {
	hermesSvc hermes.Service
}

type webhookPayload struct {
	PushData struct {
		PushedAt int64         `json:"pushed_at"`
		Images   []interface{} `json:"images"`
		Tag      string        `json:"tag"`
		Pusher   string        `json:"pusher"`
	} `json:"push_data"`
	CallbackURL string `json:"callback_url"`
	Repository  struct {
		Status          string `json:"status"`
		Description     string `json:"description"`
		IsTrusted       bool   `json:"is_trusted"`
		FullDescription string `json:"full_description"`
		RepoURL         string `json:"repo_url"`
		Owner           string `json:"owner"`
		IsOfficial      bool   `json:"is_official"`
		IsPrivate       bool   `json:"is_private"`
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		StarCount       int    `json:"star_count"`
		CommentCount    int    `json:"comment_count"`
		DateCreated     int64  `json:"date_created"`
		Dockerfile      string `json:"dockerfile"`
		RepoName        string `json:"repo_name"`
	} `json:"repository"`
}

// NewDockerHub new dockerhub handler
func NewDockerHub(svc hermes.Service) *DockerHub {
	return &DockerHub{svc}
}

func constructEventURI(payload *webhookPayload, account string) string {
	uri := fmt.Sprintf("registry:dockerhub:%s:%s:push", payload.Repository.Namespace, payload.Repository.Name)
	if account != "" {
		uri = fmt.Sprintf("%s:%s", uri, account)
	}
	return uri
}

// HandleWebhook handle DockerHub webhook
func (d *DockerHub) HandleWebhook(c *gin.Context) {
	log.Debug("Got Docker Hub webhook event")
	payload := webhookPayload{}
	if err := c.BindJSON(&payload); err != nil {
		log.WithError(err).Error("Failed to bind payload JSON to expected structure")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	event.Variables["namespace"] = payload.Repository.Namespace
	event.Variables["name"] = payload.Repository.Name
	event.Variables["tag"] = payload.PushData.Tag
	event.Variables["pusher"] = payload.PushData.Pusher
	event.Variables["pushed_at"] = time.Unix(int64(payload.PushData.PushedAt), 0).Format(time.RFC3339)

	// get secret from URL query
	event.Secret = c.Query("secret")

	// invoke trigger
	err = d.hermesSvc.TriggerEvent(eventURI, event)
	if err != nil {
		log.WithError(err).Error("Failed to trigger event pipelines")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
