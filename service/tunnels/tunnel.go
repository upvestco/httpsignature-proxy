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

package tunnels

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

	colorjson "github.com/neilotoole/jsoncolor"
	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/ui"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/rand"
)

var errTunnelNotAvailable = errors.New("events tunnel is not available")

func serviceIsNotAccessible(code int) bool {
	return code == http.StatusNotFound || code == http.StatusServiceUnavailable || code == http.StatusBadGateway
}

var AnonUserCredentials = UserCredentials{
	ClientID:     "00000000-0000-0000-0000-000000000000",
	ClientSecret: "",
}

type tunnel struct {
	apiClient    ApiClient
	eventsFilter map[string]interface{}
	logger       logger.Logger
	logHeaders   bool
	cancel       context.CancelFunc
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
	if serviceIsNotAccessible(code) {
		return errTunnelNotAvailable
	}
	if code != http.StatusOK {
		return errors.New("unexpected http response code: " + strconv.Itoa(code))
	}

	for _, item := range items {
		ui.AddPayload(item)
		if ui.IsCreated() {
			continue
		}
		formatted, origLen, filteredLen := e.filterAndFormat(item.Payload)
		if filteredLen == 0 {
			continue
		}
		filtered := origLen != filteredLen

		e.logger.PrintLn(cyan("== new webhook event received == "))
		e.logger.PrintLn(cyan("== received at: %s", item.CreatedAt.Format(time.DateTime)))
		if e.logHeaders {
			e.printHeaders(item, filtered)
		}
		payloadMessage := ""
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
	var in ui.Payload
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

func (e *tunnel) printHeaders(item ui.PullItem, filtered bool) {
	e.logger.PrintLn(cyan("== headers"))
	maxL := 0
	for key := range item.Headers {
		if l := len(key); l > maxL {
			maxL = l
		}
	}
	for key, values := range item.Headers {
		e.logger.PrintF("%s : %s", cyan("%s%s ", strings.Repeat(" ", maxL-len(key)), key), strings.Join(values, ","))
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

func (e *tunnel) filterPayload(in ui.Payload) ui.Payload {
	var out ui.Payload
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

	if err := e.apiClient.TunnelIsReady(ctx); err != nil {
		if errors.Is(err, errTunnelNotAvailable) {
			return errors.Wrap(errTunnelNotAvailable, "Webhook events listening is not available")
		}
		return errors.Wrap(err, "Could not create tunnel.")
	}

	if err := e.apiClient.Authorise(ctx, requiredScopes); err != nil {
		e.logger.PrintLn(lightRed("Could not open the Webhook events tunnel. You client must have '" + requiredScopes + "' scope(s)"))
		return nil
	}
	e.logger.LogF("client is authorised with '" + requiredScopes + "' scope(s)")

	endpoint, endpointID, err := e.apiClient.OpenEndpoint(ctx)
	if err != nil {
		return errors.Wrap(err, "Could not create tunnel.")
	}
	e.logger.LogF("backend endpoint (%s) for the client is created", endpoint)

	request := WebhookRequest{
		Title: "http signature webhook " + randomString(8),
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

	return errors.Wrap(poolErr, "doPulling on webhook "+webhookID)
}

func (e *tunnel) destroy() {
	e.cancel()
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
