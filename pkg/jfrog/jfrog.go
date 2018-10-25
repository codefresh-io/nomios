package jfrog

import (
	"encoding/json"
	"fmt"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// JFrog struct
type JFrog struct {
	hermesSvc hermes.Service
}

type webhookPayload struct {
	Artifactory struct {
		Webhook struct {
			Event string `json:event`
			Data  struct {
				Docker struct {
					Tag   string `json:tag`
					Image string `json:image`
				} `json:docker`
				Event struct {
					ModifiedBy string `json:modifiedBy`
					Created    int64  `json:created`
					RepoPath   struct {
						RepoKey string `json:repoKey`
					} `json:repoPath`
				} `json:event`
			} `json:data`
		} `json:"webhook"`
	} `json:"artifactory"`
}

// NewJFrog new jfrog handler
func NewJFrog(svc hermes.Service) *JFrog {
	return &JFrog{svc}
}

func constructEventURI(payload *webhookPayload, account string) string {
	uri := fmt.Sprintf("registry:jfrog:%s:%s:push", payload.Artifactory.Webhook.Data.Event.RepoPath.RepoKey, payload.Artifactory.Webhook.Data.Docker.Image)
	if account != "" {
		uri = fmt.Sprintf("%s:%s", uri, account)
	}
	return uri
}

// HandleWebhook handle JFrog webhook
func (d *JFrog) HandleWebhook(c *gin.Context) {
	log.Debug("Got JFrog webhook event")

	payload := webhookPayload{}
	if err := c.BindJSON(&payload); err != nil {
		log.WithError(err).Error("Failed to bind payload JSON to expected structure")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Artifactory.Webhook.Event != "docker.tagCreated" {
		log.Debug(fmt.Sprintf("Skip event %s", payload.Artifactory.Webhook.Event))
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
	event.Variables["event"] = payload.Artifactory.Webhook.Event
	event.Variables["namespace"] = payload.Artifactory.Webhook.Data.Event.RepoPath.RepoKey
	event.Variables["name"] = payload.Artifactory.Webhook.Data.Docker.Image
	event.Variables["tag"] = payload.Artifactory.Webhook.Data.Docker.Tag
	event.Variables["pusher"] = payload.Artifactory.Webhook.Data.Event.ModifiedBy
	event.Variables["pushed_at"] = time.Unix(int64(payload.Artifactory.Webhook.Data.Event.Created/1000), 0).Format(time.RFC3339)

	// get secret from URL query
	event.Secret = c.Query("secret")

	log.Debug("Event url " + eventURI)

	// invoke trigger
	err = d.hermesSvc.TriggerEvent(eventURI, event)
	if err != nil {
		log.WithError(err).Error("Failed to trigger event pipelines")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
