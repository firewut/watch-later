package config

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"net/url"
	"os"
	"project/modules/args"
	"reflect"
)

// Configuration structure
type Config struct {
	Mode          string   `flag:"mode" json:"mode"`
	HttpHost      string   `flag:"http-host" json:"http-host"`
	HostName      string   `flag:"hostname" json:"hostname"`
	HttpDomain    string   `flag:"http-domain" json:"http-domain"`
	HttpDomainURL *url.URL `flag:"-" json:"-"`
	HttpApiPath   string   `flag:"http-api-path" json:"http-api-path"`

	// hashes
	HashKey  string `flag:"hash-key" json:"-"`
	BlockKey string `flag:"block-key" json:"-"`

	AccessLog string `flag:"access-log" json:"-"`
	InfoLog   string `flag:"info-log" json:"-"`
	ErrorLog  string `flag:"error-log" json:"-"`

	TemplatesDir string `flag:"templates-dir" json:"-"`

	// Attached services
	// - deform.io
	DeformIOProject string `flag:"deformio-project" json:"-"`
	DeformIOToken   string `flag:"deformio-token" json:"-"`

	// - oauth2
	OAUTH2ClientID     string `flag:"oauth2-client-id" json:"-"`
	OAUTH2ClientSecret string `flag:"oauth2-client-secret" json:"-"`
	OAUTH2RedirectUri  string `flag:"oauth2-redirect-uri" json:"-"`
	OAUTH2AuthUri      string `flag:"oauth2-auth-uri" json:"-"`
	OAUTH2TokenUri     string `flag:"oauth2-token-uri" json:"-"`

	MiscOptions map[string]interface{} `flag:"service-misc-options" json:"-"`
	Router      *mux.Router

	SentryDSN   string        `flag:"sentry-dsn" json:"-"`
	RavenClient *raven.Client `json:"-"`
}

func NewConfig() (config *Config, err error) {
	config = &Config{}
	flags := map[string]string{}
	for k, v := range args.Flags {
		flags[k] = v
	}
	config.ApplyOptions(config, flags)

	var u *url.URL
	u, err = url.Parse(config.HttpDomain)
	if err == nil {
		if len(u.Scheme) == 0 {
			err = errors.New("Need a http/https scheme")
		}
		if len(u.Host) == 0 {
			err = errors.New("Need a host")
		}
		if err == nil {
			config.HttpDomainURL = u
		}
	}

	if len(config.SentryDSN) > 0 {
		config.RavenClient, err = raven.New(config.SentryDSN)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("Need a sentry-dsn"))
	}

	return
}

func (c *Config) ApplyOptions(config *Config, config_request_content interface{}) {
	configs_map := config_request_content.(map[string]string)
	if len(configs_map) > 0 {
		misc_options := map[string]interface{}{}
		if len(config.MiscOptions) > 0 {
			misc_options = config.MiscOptions
		}
		for k, v := range configs_map {
			s := reflect.ValueOf(config).Elem()
			for i := 0; i < s.NumField(); i++ {
				typeField := s.Type().Field(i)
				field := s.FieldByName(typeField.Name)
				tag := typeField.Tag
				flag_tag := tag.Get("flag")

				os_val := os.Getenv(flag_tag)
				if len(os_val) > 0 {
					field.Set(reflect.ValueOf(os_val))
				}

				if value, ok := configs_map[flag_tag]; ok {
					field.Set(reflect.ValueOf(value))
				} else {
					misc_options[k] = v
				}
			}
		}
		config.MiscOptions = misc_options
	}
}
