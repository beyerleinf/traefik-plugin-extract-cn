package traefik_plugin_extract_cn

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
)

type Config struct {
	DestHeader string `json:"dest,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type ExtractClientCert struct {
	next       http.Handler
	destHeader string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.DestHeader == "" {
		return nil, errors.New("destHeader must be specified")
	}

	return &ExtractClientCert{
		next:       next,
		destHeader: config.DestHeader,
	}, nil
}

func (e *ExtractClientCert) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	certInfo := req.Header.Get("X-Forwarded-Tls-Client-Cert-Info")
	if certInfo != "" {
		decoded, err := url.QueryUnescape(certInfo)
		if err == nil {
			re := regexp.MustCompile(`CN=([^,/"]+)`)
			if matches := re.FindStringSubmatch(decoded); len(matches) > 1 {
				req.Header.Set(e.destHeader, matches[1])
			}
		}
	}

	e.next.ServeHTTP(rw, req)
}
