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
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gookit/color"
	colorjson "github.com/neilotoole/jsoncolor"
	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"golang.org/x/exp/maps"
)

type tunnel struct {
	apiClient    ApiClient
	eventsFilter map[string]interface{}
	logger       logger.Logger
	logHeaders   bool
	cancel       context.CancelFunc
}

var cyan = fmt.Sprintf
var lightRed = fmt.Sprintf

func init() {
	if colorjson.IsColorTerminal(os.Stdout) {
		cyan = color.FgCyan.Sprintf
		lightRed = color.FgLightRed.Sprintf
	}
}

func createTunnel(apiClient ApiClient, events []string, logHeaders bool, logger logger.Logger) *tunnel {
	eventsFilter := map[string]interface{}{}
	for _, t := range events {
		if len(t) == 0 {
			continue
		}
		eventsFilter[t] = 1
	}

	return &tunnel{
		apiClient:    apiClient,
		eventsFilter: eventsFilter,
		logger:       logger,
		logHeaders:   logHeaders,
	}
}

const requiredScopes = "webhooks:admin"

func (e *tunnel) doPulling(ctx context.Context, endpointID string) error {

	e.logger.LogF("start pulling evens")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := e.pullEvents(ctx, endpointID); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				if errors.Is(err, syscall.ECONNREFUSED) {
					return nil
				}
				return errors.Wrap(err, "pullEvents")
			}
		}
	}
}

func (e *tunnel) pullEvents(ctx context.Context, endpointID string) error {
	items, code, err := e.apiClient.GetEvents(ctx, endpointID)
	if err != nil {
		return errors.Wrap(err, "doPull")
	}
	if code == http.StatusUnauthorized {
		if err := e.apiClient.Authorise(ctx, requiredScopes); err != nil {
			return errors.Wrap(err, "Could not open the Webhook events tunnel. You client must have '"+requiredScopes+"' scope(s)")
		}
		items, code, err = e.apiClient.GetEvents(ctx, endpointID)
		if err != nil {
			return errors.Wrap(err, "doPull")
		}
	}
	if code != http.StatusOK {
		return errors.New("unexpected http response code: " + strconv.Itoa(code))
	}

	for _, item := range items {
		formatted, origLen, filteredLen := e.filterAndFormat(item.Payload)
		if filteredLen == 0 {
			continue
		}
		filtered := origLen != filteredLen

		e.logger.PrintLn(cyan("== new incoming webhook request"))
		e.logger.PrintLn(cyan("== received at: %s", item.CreatedAt.Format(time.DateTime)))
		if e.logHeaders {
			e.printHeaders(item, filtered)
		}
		payloadMessage := "== payload"
		if filtered {
			payloadMessage += fmt.Sprintf(" was filtered: origin events: %d, filtered events: %d)", origLen, filteredLen)
		}
		e.logger.PrintLn(cyan(payloadMessage))
		e.logger.PrintLn(formatted)
	}
	return nil
}

func (e *tunnel) filterAndFormat(payload string) (string, int, int) {
	if len(payload) == 0 {
		return "", 0, 0
	}
	var in Payload
	if err := json.Unmarshal([]byte(payload), &in); err != nil {
		return payload, 1, 1
	}
	out := e.filterPayload(in)

	buff := &bytes.Buffer{}
	encoder := colorjson.NewEncoder(buff)
	encoder.SetIndent("", " ")
	if colorjson.IsColorTerminal(os.Stdout) {
		encoder.SetColors(colorjson.DefaultColors())
	}
	_ = encoder.Encode(out)

	return buff.String(), len(in.Payload), len(out.Payload)
}

func (e *tunnel) printHeaders(item PullItem, filtered bool) {
	e.logger.PrintLn(cyan("== headers"))
	maxL := 0
	for key := range item.Headers {
		if l := len(key); l > maxL {
			maxL = l
		}
	}
	for key, values := range item.Headers {
		e.logger.Print(cyan("%s%s ", strings.Repeat(" ", maxL-len(key)), key))
		e.logger.PrintF(": %s", strings.Join(values, ","))
		remarks := ""
		if key == "Content-Length" {
			remarks = " # The Content-Length header shows the length of the original payload. The payload was formated"
			if filtered {
				remarks += " and filtered by events filter"
			}
			remarks += "."
		}
		e.logger.PrintLn(lightRed(remarks))
	}
}

func (e *tunnel) filterPayload(in Payload) Payload {
	var out Payload
	if len(e.eventsFilter) == 0 {
		return in
	}
	for _, event := range in.Payload {
		if _, exists := e.eventsFilter[event.Type]; exists {
			out.Payload = append(out.Payload, event)
		}
	}
	return out
}

func (e *tunnel) start() error {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	if err := e.apiClient.Authorise(ctx, requiredScopes); err != nil {
		return errors.Wrap(err, "Could not open the Webhook events tunnel. You client must have '"+requiredScopes+"' scope(s)")
	}
	e.logger.LogF("client is authorised with '" + requiredScopes + "' scope(s)")

	endpoint, endpointID, err := e.apiClient.OpenEndpoint(ctx)
	if err != nil {
		return errors.Wrap(err, "Could not create tunnel.")
	}
	e.logger.LogF("backend endpoint (%s) for the client is created", endpoint)

	request := WebhookRequest{
		Title: "http signature temporary webhook",
		Url:   endpoint,
		Type:  []string{"ALL"},
		Config: &WebhookConfig{
			Delay:          "1s",
			MaxPackageSize: 12400,
		},
	}

	webhookID, err := e.apiClient.CreateWebhook(ctx, request)
	if err != nil {
		return errors.Wrap(err, "Could not create webhook")
	}

	request.Enabled = true
	if err := e.apiClient.PatchWebhook(ctx, webhookID, request); err != nil {
		return errors.Wrap(err, "Could not enable webhook")
	}
	events := "ALL"
	if len(e.eventsFilter) > 0 {
		events = strings.Join(maps.Keys(e.eventsFilter), ",")
	}
	e.logger.PrintLn(cyan("Listen for the [%s] events by the webhook %s", events, webhookID))

	poolErr := e.doPulling(ctx, endpointID)

	ctx = context.Background()
	if err := e.apiClient.DeleteWebhook(ctx, webhookID); err != nil {
		if !errors.Is(err, syscall.ECONNREFUSED) {
			e.logger.LogF("Fail to delete webhook: %s", err.Error())
		}
	}
	if err := e.apiClient.CloseEndpoint(ctx, endpointID); err != nil {
		if !errors.Is(err, syscall.ECONNREFUSED) {
			e.logger.LogF("Fail to close endpoint: %s", err.Error())
		}
	}

	return errors.Wrap(poolErr, "doPulling")
}

func (e *tunnel) destroy() {
	e.cancel()
}

type Payload struct {
	Payload []struct {
		CreatedAt time.Time              `json:"created_at"`
		Id        string                 `json:"id"`
		Object    map[string]interface{} `json:"object"`
		Type      string                 `json:"type"`
		WebhookId string                 `json:"webhook_id"`
	} `json:"payload"`
}
