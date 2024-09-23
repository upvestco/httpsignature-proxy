/*
Copyright © 2021 Upvest GmbH <support@upvest.co>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/valyala/fastjson"
)

type ApiClient interface {
	Authorise(context.Context, string) error
	CreateWebhook(context.Context, WebhookRequest) (string, error)
	PatchWebhook(context.Context, string, WebhookRequest) error
	DeleteWebhook(context.Context, string) error
	OpenEndpoint(context.Context) (string, string, error)
	CloseEndpoint(context.Context, string) error
	GetEvents(context.Context, string) ([]PullItem, int, error)
}

func NewClient(proxyAddress string, usersCredentials UserCredentials, timeout time.Duration) ApiClient {
	return &apiClient{
		proxyAddress:     proxyAddress,
		usersCredentials: usersCredentials,
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

type apiClient struct {
	proxyAddress     string
	usersCredentials UserCredentials
	accessToken      string
	httpClient       http.Client
}

func (e *apiClient) Authorise(ctx context.Context, scopes string) error {
	authReq := []byte(fmt.Sprintf(`client_id=%s&client_secret=%s&grant_type=client_credentials&scope=%s`,
		e.usersCredentials.ClientID, e.usersCredentials.ClientSecret, scopes))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.proxyAddress+"/auth/token", bytes.NewBuffer(authReq))
	if err != nil {
		return errors.Wrap(err, "NewRequestWithContext")
	}
	req.Header.Add("Upvest-Client-Id", e.usersCredentials.ClientID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add(logger.HttpProxyNoLogging, "true")

	code, body, err := e.io(req)

	if err != nil {
		return errors.Wrap(err, "io")
	}
	if code != http.StatusOK {
		return errors.New("Wrong http code: " + strconv.Itoa(code))
	}
	accessToken := fastjson.GetString(body, "access_token")
	if len(accessToken) == 0 {
		return errors.New("no access token")
	}
	e.accessToken = accessToken
	return nil
}

// WebhookRequest see: https://docs.upvest.co/documentation/getting_started/implementing_webhooks/webhook_registration
type WebhookRequest struct {
	Title   string         `json:"title"`
	Url     string         `json:"url"`
	Type    []string       `json:"type"`
	Enabled bool           `json:"enabled,omitempty"`
	Config  *WebhookConfig `json:"config,omitempty"`
}
type WebhookConfig struct {
	Delay          string `json:"delay,omitempty"`
	MaxPackageSize int    `json:"max_package_size,omitempty"`
}

func (e *apiClient) PatchWebhook(ctx context.Context, id string, webhookRequest WebhookRequest) error {
	return e.patch(ctx, webhookRequest, "/webhooks/"+id)
}

func (e *apiClient) CreateWebhook(ctx context.Context, webhookRequest WebhookRequest) (string, error) {
	body, err := e.create(ctx, webhookRequest, "/webhooks")
	if err != nil {
		return "", errors.Wrap(err, "create")
	}
	id := fastjson.GetString(body, "id")
	if len(id) == 0 {
		return "", errors.New("no webhook id")
	}
	return id, nil
}

type PullItem struct {
	Headers   http.Header `json:"headers"`
	Payload   string      `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

func (e *apiClient) GetEvents(ctx context.Context, tunnelID string) ([]PullItem, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.proxyAddress+"/events-acceptor-service/endpoints/"+tunnelID, nil)
	if err != nil {
		return nil, 0, errors.Wrap(err, "NewRequestWithContext")
	}
	e.addHeaders(req)
	code, body, err := e.io(req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "io")
	}
	var res []PullItem
	if code == http.StatusOK {
		if err := json.Unmarshal(body, &res); err != nil {
			return nil, 0, errors.Wrap(err, "Unmarshal")
		}
	}
	return res, code, nil
}

func (e *apiClient) OpenEndpoint(ctx context.Context) (string, string, error) {
	body, err := e.create(ctx, "", "/events-acceptor-service/endpoints")
	if err != nil {
		return "", "", errors.Wrap(err, "create")
	}
	url := fastjson.GetString(body, "url")
	if len(url) == 0 {
		return "", "", errors.New("no tunnel url")
	}
	id := fastjson.GetString(body, "id")
	if len(id) == 0 {
		return "", "", errors.New("no tunnel id")
	}
	return url, id, nil
}

func (e *apiClient) CloseEndpoint(ctx context.Context, tunnelID string) error {
	return e.delete(ctx, "/events-acceptor-service/endpoints/"+tunnelID)
}

func (e *apiClient) DeleteWebhook(ctx context.Context, id string) error {
	return e.delete(ctx, "/webhooks/"+id)
}

func (e *apiClient) delete(ctx context.Context, uri string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, e.proxyAddress+uri, nil)
	if err != nil {
		return errors.Wrap(err, "NewRequestWithContext")
	}
	e.addHeaders(req)
	code, _, err := e.io(req)
	if err != nil {
		return errors.Wrap(err, "io")
	}
	if code != http.StatusNoContent {
		return errors.New("Wrong http code: " + strconv.Itoa(code))
	}
	return nil
}

func (e *apiClient) create(ctx context.Context, request interface{}, uri string) ([]byte, error) {
	out, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Marshal")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.proxyAddress+uri, bytes.NewBuffer(out))
	if err != nil {
		return nil, errors.Wrap(err, "NewRequestWithContext")
	}
	e.addHeaders(req)
	code, body, err := e.io(req)
	if err != nil {
		return nil, errors.Wrap(err, "io")
	}
	if code != http.StatusCreated {
		return nil, errors.New("Wrong http code: " + strconv.Itoa(code))
	}
	return body, err
}

func (e *apiClient) patch(ctx context.Context, request interface{}, uri string) error {
	out, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, e.proxyAddress+uri, bytes.NewBuffer(out))
	if err != nil {
		return errors.Wrap(err, "NewRequestWithContext")
	}
	e.addHeaders(req)
	code, _, err := e.io(req)
	if err != nil {
		return errors.Wrap(err, "io")
	}
	if code != http.StatusOK {
		return errors.New("Wrong http code: " + strconv.Itoa(code))
	}
	return nil
}

func (e *apiClient) io(req *http.Request) (int, []byte, error) {
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return 0, nil, errors.Wrap(err, "httpClient.Do")
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, errors.Wrap(err, "ReadAll")
	}
	return resp.StatusCode, body, nil
}

func (e *apiClient) addHeaders(req *http.Request) {
	req.Header.Add(upvestClientID, e.usersCredentials.ClientID)
	req.Header.Add("Authorization", "Bearer "+e.accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(logger.HttpProxyNoLogging, "true")
}
