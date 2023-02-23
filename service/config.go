package service

import (
	"github.com/vibeitco/go-utils/config"
)

type Config struct {
	config.Core
	SpaceapiAuth string `json:"spaceapiAuth" yaml:"spaceapiAuth" envconfig:"spaceapiAuth"`
}
