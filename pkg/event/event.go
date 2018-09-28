package event

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type (
	// Info event info - EVERYTHING for specific event (eventURI)
	Info struct {
		// Endpoint URL
		Endpoint string `json:"endpoint,omitempty"`
		// Description human readable text
		Description string `json:"description,omitempty"`
		// Status current event handler status (active, error, not active)
		Status string `json:"status,omitempty"`
		// Help test
		Help string `json:"help,omitempty"`
	}
)

const validURI = `^registry:(dockerhub|quay):[a-z0-9_-]+:[a-z0-9_-]+:push(:[[:xdigit:]]{12})$`

// compiled validator regexp
var validator, _ = regexp.Compile(validURI)

// GetEventInfo get extended info from uri
func GetEventInfo(publicDNS string, uri string, secret string) (*Info, error) {
	log.WithFields(log.Fields{
		"event-uri": uri,
		"validator": validURI,
	}).Debug("get trigger-event info")
	if !validator.MatchString(uri) {
		log.Error("failed to match event URI")
		return nil, fmt.Errorf("unexpected event uri: %s", uri)
	}
	// split uri
	s := strings.Split(uri, ":")
	kind := s[1]
	// dockerhub repository 1st
	repo := s[2]
	// dockerhub image 2nd
	image := s[3]
	// get account hash (may not exist)
	var account string
	if len(s) == 6 {
		account = s[5]
	}

	//TODO: refactor it, move to different classes , dockerhub_info.go and quay_info.go and imepelement different methods
	var humanReadableType string
	var settingsLink string
	if kind == "quay"{
		humanReadableType = "Quay"
		settingsLink = fmt.Sprintf("https://quay.io/repository/%s/%s?tab=settings", repo, image)
	} else {
		humanReadableType = "Docker Hub"
		settingsLink = fmt.Sprintf("https://hub.docker.com/r/%s/%s/~/settings/webhooks/", repo, image)
	}


	// format info
	info := new(Info)
	info.Description = fmt.Sprintf("%s %s/%s push event", humanReadableType, repo, image)
	// handle endpoint url
	u, err := url.Parse(publicDNS)
	if err != nil {
		log.WithError(err).WithField("dns", publicDNS).Warn("failed to parse public dns")
	} else {
		q := u.Query()
		q.Set("secret", secret)
		if account != "" {
			q.Set("account", account)
		}
		u.Path = "nomios/" + kind + u.Path
		u.RawQuery = q.Encode()
		info.Endpoint = u.String()
	}
	info.Status = "active"
	info.Help = fmt.Sprintf(`%s webhooks fire when an image is built in, pushed or a new tag is added to, your repository.

Configure %s on %s

Add following Codefresh %s webhook endpoint %s`, humanReadableType, humanReadableType, settingsLink, humanReadableType, info.Endpoint)

	// return info
	return info, nil
}

// Subscribe to event in DockerHub
func Subscribe(publicDNS, uri, secret, credentials string) (*Info, error) {
	return nil, nil
}

// Unsubscribe from event in DockerHub
func Unsubscribe(publicDNS, uri, credentials string) (*Info, error) {
	return nil, nil
}
