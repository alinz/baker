package container_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/alinz/baker"
	"github.com/alinz/baker/container"
)

func mockResponse(t *testing.T, scenario string, r *http.Request) io.ReadCloser {
	file, err := os.Open(path.Join("./fixtures/docker/api", scenario, r.URL.String(), "payload.json"))
	if err != nil {
		t.Fatal(err)
	}

	return file
}

func mockDockerServer(t *testing.T, scenario string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockResponse(t, scenario, r)
		defer resp.Close()

		w.WriteHeader(http.StatusOK)
		io.Copy(w, resp)
	}))

	return server
}

type DummyConsumer struct {
	container func(container *baker.Container) error
	close     func(err error)
}

var _ (container.Consumer) = (*DummyConsumer)(nil)

func (dc *DummyConsumer) Container(container *baker.Container) error {
	return dc.container(container)
}

func (dc *DummyConsumer) Close(err error) {
	dc.close(err)
}

func TestDocker(t *testing.T) {

	testCases := []struct {
		scenario string
	}{
		{
			scenario: "scenario1",
		},
	}

	for _, testCase := range testCases {
		server := mockDockerServer(t, testCase.scenario)
		defer server.Close()

		docker := container.NewDocker(server.Client(), server.URL)
		docker.Pipe(&DummyConsumer{
			container: func(container *baker.Container) error {
				fmt.Println(container)
				return nil
			},
			close: func(err error) {
				fmt.Println(err)
			},
		})
	}
}
