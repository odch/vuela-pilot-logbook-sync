package capzlog

import (
	"log"
	"os"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/odch/go-capzlog/client"
)

// BearerToken provides a header based oauth2 bearer access token auth info writer
func BearerBasicToken(token, systemInstanceIdentifier string) runtime.ClientAuthInfoWriter {
	return runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("SystemInstanceIdentifier", systemInstanceIdentifier)

		return r.SetHeaderParam(runtime.HeaderAuthorization, "Basic "+token)
	})
}

var c *client.Capzlog

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	host := getEnv("CAPZLOG_HOST", client.DefaultHost)
	log.Println(host)
	dc := client.DefaultTransportConfig().WithHost(host)
	c = client.NewHTTPClientWithConfig(nil, dc)
}
