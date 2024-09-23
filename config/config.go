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

package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	DefaultClientKey = "default"
)

type Config struct {
	BaseConfig     *BaseConfig
	KeyConfigs     []KeyConfig
	DefaultTimeout time.Duration
	PullDelay      time.Duration
	VerboseMode    bool
	LogHeaders     bool
	Port           int
}

type BaseConfig struct {
	BaseUrl            string
	KeyID              string
	PrivateKeyFileName string
	Password           string
}

type KeyConfig struct {
	ClientID string
	BaseConfig
}

func (c *BaseConfig) IsEmpty() bool {
	return c.BaseUrl == "" &&
		c.KeyID == "" && c.Password == "" && c.PrivateKeyFileName == ""
}

func (c *BaseConfig) Validate() error {
	if c.KeyID == "" {
		return errors.New("keyID is empty")
	}
	if _, err := os.Stat(c.PrivateKeyFileName); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("private key file not exists: %s", c.PrivateKeyFileName)
	}
	if _, err := url.Parse(c.BaseUrl); err != nil || c.BaseUrl == "" {
		return errors.New("base url is empty or invalid")
	}
	return nil
}

func (c *KeyConfig) IsEmpty() bool {
	return c.BaseConfig.IsEmpty() && c.ClientID == ""
}

func (c *KeyConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	if c.ClientID == DefaultClientKey {
		return nil
	}
	_, err := uuid.Parse(c.ClientID)
	if err != nil {
		return errors.New("clientID is not a valid uuid")
	}
	return nil
}
