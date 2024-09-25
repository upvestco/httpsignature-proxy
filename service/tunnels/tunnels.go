package tunnels

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/gookit/color"
	colorjson "github.com/neilotoole/jsoncolor"
	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/ui"
	"golang.org/x/exp/maps"
)

var cyan = fmt.Sprintf
var lightRed = fmt.Sprintf

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
	if !ui.IsCreated() {
		if colorjson.IsColorTerminal(os.Stdout) {
			cyan = color.FgCyan.Sprintf
			lightRed = color.FgLightRed.Sprintf
		}
	}
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

func (e *Tunnels) Start(userCredentialsCh chan UserCredentials) {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	if err := e.createApiClient(AnonUserCredentials).TunnelIsReady(ctx); err != nil {
		if errors.Is(err, errTunnelNotAvailable) {
			e.logger.PrintLn(cyan("Webhook events listening is not available"))
		} else {
			e.logger.Log(err.Error())
		}
		return
	}

	e.logger.PrintLn(cyan("###############################################################"))
	e.logger.PrintLn(cyan("To start event listener, send an auth request: POST /auth/token"))
	e.logger.PrintLn(cyan("###############################################################"))

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
					e.logger.PrintLn(lightRed(err.Error()))
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
