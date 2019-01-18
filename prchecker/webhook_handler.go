package prchecker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// WebhookHandler handles webhook request.
type WebhookHandler struct {
	client *github.Client
	Config Config
}

// New creates WebhookHandler from environment params.
func New() (*WebhookHandler, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return NewWithConfig(conf)
}

// NewWithConfig creates WebhookHandler from given Config.
func NewWithConfig(conf Config) (*WebhookHandler, error) {
	h := &WebhookHandler{}

	cli, err := newClient(conf)
	if err != nil {
		h.loggingError("newClient error: %s", err)
		return nil, err
	}

	if conf.BotID == 0 {
		// get bot's id from github's current user.
		ctx := context.Background()
		u, _, err := cli.Users.Get(ctx, "")
		if err != nil {
			return nil, err
		}

		conf.BotID = u.GetID()
		h.loggingInfo("botID is fetched from github: %d", conf.BotID)
	}

	h.client = cli
	h.Config = conf
	h.loggingInfo("initialized")
	return h, nil
}

// newClient creates github.Client from given Config.
func newClient(conf Config) (*github.Client, error) {
	ctx := context.Background()

	var ts oauth2.TokenSource
	switch {
	case conf.HasAPIToken():
		ts = oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: conf.GetAPIToken()},
		)
	default:
		return nil, errors.New("Cannot find GitHub credentials")
	}

	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
}

// HandleRequest handles webhook request and run PullRequestChecker.
func (h *WebhookHandler) HandleRequest(r *http.Request) error {
	payload, err := h.getPayload(r)
	if err != nil {
		h.loggingError("payload parse error: %s", err)
		return err
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		h.loggingError("github.ParseWebHook error: %s", err)
		return err
	}

	switch e := event.(type) {
	case *github.PullRequestEvent:
		pushData := github.PushEvent{}
		json.Unmarshal(payload, &pushData)
		checker := newPullRequestChecker(h.Config, h.client, e, payload)
		return checker.Do()
	default:
		return nil
	}
}

// getPayload extract payload data from Request.
func (h *WebhookHandler) getPayload(r *http.Request) ([]byte, error) {
	switch {
	case h.Config.WebHookSecret != "":
		return github.ValidatePayload(r, []byte(h.Config.WebHookSecret))
	default:
		return ioutil.ReadAll(r.Body)
	}
}

func (h *WebhookHandler) loggingError(template string, err error) {
	fmt.Printf("[WebhookHandler] [ERROR] [botID:%d] %s\n", h.Config.BotID, fmt.Sprintf(template, err.Error()))
}

func (h *WebhookHandler) loggingInfo(template string, params ...interface{}) {
	fmt.Printf("[WebhookHandler] [INFO] [botID:%d] %s\n", h.Config.BotID, fmt.Sprintf(template, params...))
}
