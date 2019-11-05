package endpoint_test

import (
	"testing"

	"github.com/alinz/baker/pkg/endpoint"
)

func TestAddr(t *testing.T) {
	testCases := []struct {
		host   string
		port   int
		secure bool

		expected string
	}{
		{
			host:   "0.0.0.0",
			port:   1,
			secure: false,

			expected: "0.0.0.0:1",
		},
		{
			host:   "0.0.0.0",
			port:   1,
			secure: true,

			expected: "0.0.0.0:1",
		},
	}

	for _, testCase := range testCases {
		addr := endpoint.NewAddr(testCase.host, testCase.port, testCase.secure)
		if addr.String() != testCase.expected {
			t.Errorf("expected '%s' but got '%s'", addr.String(), testCase.expected)
		}

		if addr.Host() != testCase.host {
			t.Errorf("expected host to be '%s' but got '%s'", addr.Host(), testCase.host)
		}

		if addr.Port() != testCase.port {
			t.Errorf("expected port to be '%d' but got '%d'", addr.Port(), testCase.port)
		}

		if addr.Secure() != testCase.secure {
			t.Errorf("expected secure to be '%t' but got '%t'", addr.Secure(), testCase.secure)
		}
	}
}

func TestHTTPAddr(t *testing.T) {
	testCases := []struct {
		host   string
		port   int
		secure bool
		path   string

		expected string
	}{
		{
			host:   "0.0.0.0",
			port:   1,
			secure: false,
			path:   "/",

			expected: "http://0.0.0.0:1/",
		},
		{
			host:   "0.0.0.0",
			port:   1,
			secure: true,
			path:   "/",

			expected: "https://0.0.0.0:1/",
		},
		{
			host:   "0.0.0.0",
			port:   1,
			secure: false,
			path:   "/hello",

			expected: "http://0.0.0.0:1/hello",
		},
		{
			host:   "0.0.0.0",
			port:   1,
			secure: true,
			path:   "/hello",

			expected: "https://0.0.0.0:1/hello",
		},
		{
			host:   "localhost",
			port:   80,
			secure: false,
			path:   "",

			expected: "http://localhost:80/",
		},
	}

	for _, testCase := range testCases {
		addr := endpoint.NewAddr(testCase.host, testCase.port, testCase.secure)
		httpAddr := endpoint.NewHTTPAddr(addr, testCase.path)
		if httpAddr.String() != testCase.expected {
			t.Errorf("expected '%s' but got '%s'", httpAddr.String(), testCase.expected)
		}

		if httpAddr.Host() != testCase.host {
			t.Errorf("expected host to be '%s' but got '%s'", httpAddr.Host(), testCase.host)
		}

		if httpAddr.Port() != testCase.port {
			t.Errorf("expected port to be '%d' but got '%d'", httpAddr.Port(), testCase.port)
		}

		if httpAddr.Secure() != testCase.secure {
			t.Errorf("expected secure to be '%t' but got '%t'", httpAddr.Secure(), testCase.secure)
		}

		if httpAddr.Path() != testCase.path {
			t.Errorf("expected path to be '%s' but got '%s'", httpAddr.Path(), testCase.path)
		}
	}
}

func TestParseHTTP(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{
			url: "http://localhost:9000/hello",
		},
		{
			url: "http://localhost:80/",
		},
		{
			url: "http://localhost:9000/hello",
		},
		{
			url: "https://localhost:9000/hello",
		},
	}

	for _, testCase := range testCases {
		u := endpoint.ParseHTTPAddr(testCase.url)
		if u.String() != testCase.url {
			t.Fatalf("expected '%s' but got '%s'", testCase.url, u.String())
		}
	}
}
