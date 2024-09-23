package runtime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"golang.org/x/exp/maps"
)

type UserCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (e UserCredentials) Empty() bool {
	return len(e.ClientSecret) == 0 && len(e.ClientID) == 0
}

type Tunnels struct {
	logger          logger.Logger
	events          []string
	closeGroup      *sync.WaitGroup
	cancel          context.CancelFunc
	tunnels         *tunnelsMap
	logHeaders      bool
	createApiClient func(credentials UserCredentials) ApiClient
	proxyAddress    string
}

func CreateTunnels(logger logger.Logger, events []string, proxyAddress string, createApiClient func(credentials UserCredentials) ApiClient, logHeaders bool) *Tunnels {
	return &Tunnels{
		logger:          logger,
		tunnels:         newTunnelsMap(),
		events:          events,
		closeGroup:      new(sync.WaitGroup),
		createApiClient: createApiClient,
		logHeaders:      logHeaders,
		proxyAddress:    proxyAddress,
	}
}

func (e *Tunnels) Stop() {
	e.cancel()
	if list := e.tunnels.list(); len(list) > 0 {
		e.logger.Log("Closing webhooks tunnels")
		for _, t := range list {
			t.destroy()
		}
	}
	e.closeGroup.Wait()
}

func AskForUserCredentials(proxyAddress string) (UserCredentials, int, error) {
	client := http.Client{
		Timeout: time.Second,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, proxyAddress+"/proxy-pass", nil)
	if err != nil {
		return UserCredentials{}, 0, errors.Wrap(err, "NewRequestWithContext")
	}
	resp, err := client.Do(req)
	if err != nil {
		return UserCredentials{}, 0, errors.Wrap(err, "DefaultClient.Do")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusAccepted {
		return UserCredentials{}, resp.StatusCode, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserCredentials{}, 0, errors.Wrap(err, "ReadAll")
	}
	var uc UserCredentials

	if err := json.Unmarshal(body, &uc); err != nil {
		return UserCredentials{}, http.StatusBadRequest, nil
	}
	return uc, resp.StatusCode, nil
}

func (e *Tunnels) Start(userCredentialsCh chan UserCredentials) {
	e.logger.Print(cyan("############################################################\n"))
	e.logger.Print(cyan("To start listening webhook events - send /auth/token request\n"))
	e.logger.Print(cyan("############################################################\n"))
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	for {
		select {
		case <-ctx.Done():
			return
		case uc := <-userCredentialsCh:
			if uc.Empty() {
				continue
			}
			exists := e.tunnels.exists(uc.ClientID)
			if exists {
				continue
			}
			t := createTunnel(e.createApiClient(uc), e.events, e.logHeaders, e.logger)
			e.tunnels.add(uc.ClientID, t)
			e.closeGroup.Add(1)
			go func(t *tunnel, uc UserCredentials) {
				if err := t.start(); err != nil {
					e.logger.Log(err.Error() + "\nListener closed")
				}
				e.tunnels.remove(uc.ClientID)
				e.closeGroup.Done()
			}(t, uc)
		}
	}
}

type tunnelsMap struct {
	tunnels map[string]*tunnel
	lo      *sync.Mutex
}

func newTunnelsMap() *tunnelsMap {
	return &tunnelsMap{
		tunnels: map[string]*tunnel{},
		lo:      new(sync.Mutex),
	}
}
func (e *tunnelsMap) add(client string, tu *tunnel) {
	e.lo.Lock()
	e.tunnels[client] = tu
	e.lo.Unlock()
}
func (e *tunnelsMap) remove(client string) {
	e.lo.Lock()
	delete(e.tunnels, client)
	e.lo.Unlock()
}
func (e *tunnelsMap) exists(client string) bool {
	e.lo.Lock()
	_, exists := e.tunnels[client]
	e.lo.Unlock()
	return exists
}
func (e *tunnelsMap) list() []*tunnel {
	e.lo.Lock()
	list := maps.Values(e.tunnels)
	e.lo.Unlock()
	return list
}
