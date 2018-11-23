package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	l "github.com/spothero/segment-proxy/pkg/logging"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var version = "not-set"

// singleJoiningSlash is copied from httputil.singleJoiningSlash method.
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// NewSegmentReverseProxy is adapted from the httputil.NewSingleHostReverseProxy
// method, modified to dynamically redirect to different servers (CDN or Tracking API)
// based on the incoming request, and sets the host of the request to the host of of
// the destination URL.
func NewSegmentReverseProxy(cdn *url.URL, trackingAPI *url.URL) http.Handler {
	director := func(req *http.Request) {
		// Figure out which server to redirect to based on the incoming request.
		var target *url.URL
		if strings.HasPrefix(req.URL.String(), "/v1/projects") || strings.HasPrefix(req.URL.String(), "/analytics.js/v1") {
			target = cdn
		} else {
			target = trackingAPI
		}

		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		// Set the host of the request to the host of of the destination URL.
		// See http://blog.semanticart.com/blog/2013/11/11/a-proper-api-proxy-written-in-go/.
		req.Host = req.URL.Host
		logger.Infow("Processing request", "url", req.URL.String())
	}
	return &httputil.ReverseProxy{Director: director}
}

func healthResponse(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(version))
}

var port = flag.String("port", "8080", "bind address")
var healthport = flag.String("healthport", "8081", "bind address")
var debug = flag.Bool("debug", false, "debug mode")

func main() {
	flag.Parse()

	var config = l.LoggingConfig{
		Level:      "DEBUG",
		AppVersion: version,
	}

	config.InitializeLogger()
	logger = l.Logger.Sugar()
	cdnURL, err := url.Parse("https://cdn.segment.com")
	if err != nil {
		logger.Error(err)
	}
	trackingAPIURL, err := url.Parse("https://api.segment.io")
	if err != nil {
		logger.Error(err)
	}
	proxy := NewSegmentReverseProxy(cdnURL, trackingAPIURL)
	if *debug {
		proxy = handlers.LoggingHandler(os.Stdout, proxy)
	}

	shutdown := make(chan bool)

	go func() {
		healthServer := http.NewServeMux()
		healthServer.HandleFunc("/health", healthResponse)
		logger.Infof("Serving healthcheck at port %v", *healthport)
		logger.Error(http.ListenAndServe(":"+*healthport, healthServer))
		shutdown <- true
	}()

	go func() {
		logger.Infof("Serving proxy at port %v", *port)
		logger.Error(http.ListenAndServe(":"+*port, proxy))
		shutdown <- true
	}()

	// Blocks and waits until it receives a bool
	<-shutdown
}
