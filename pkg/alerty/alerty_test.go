package alerty_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/samber/lo"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/alerty-ai/alerty-go/pkg/alerty"
)

func testServer(t *testing.T, testFunc func(*collectortracepb.ExportTraceServiceRequest) bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exportTraceRequest := &collectortracepb.ExportTraceServiceRequest{}
		err := proto.Unmarshal(lo.Must(io.ReadAll(r.Body)), exportTraceRequest)
		if err != nil {
			t.Fatalf("failed to unmarshal ExportTraceServiceRequest: %v", err)
		}

		if !testFunc(exportTraceRequest) {
			t.Fatalf("testFunc failed")
		}

		w.WriteHeader(http.StatusOK)
	}))

}

func TestCaptureError(t *testing.T) {
	testServer := testServer(t, func(req *collectortracepb.ExportTraceServiceRequest) bool {
		return strings.Contains(req.String(), "test error")
		// && strings.Contains(req.String(), "TestCaptureError") // TODO: stack trace
	})

	alerty.Start(alerty.AlertyServiceConfig{
		Name:        "test",
		Version:     "1.0.0",
		Environment: "test",
		IngestURL:   testServer.URL,
	})
	defer alerty.Stop()

	err := errors.New("test error")
	alerty.CaptureError(err)
}

func TestCapturePanic(t *testing.T) {
	testServer := testServer(t, func(req *collectortracepb.ExportTraceServiceRequest) bool {
		return strings.Contains(req.String(), "local panic")
		// && strings.Contains(req.String(), "TestCapturePanic") // TODO: stack trace
	})

	alerty.Start(alerty.AlertyServiceConfig{
		Name:        "test",
		Version:     "1.0.0",
		Environment: "test",
		IngestURL:   testServer.URL,
	})
	defer alerty.Stop()

	doPanic := func() {
		defer func() {
			if r := recover(); r != nil {
				alerty.CapturePanic(r)
			}
		}()

		panic("local panic")
	}

	doPanic()
}

func TestRecoverHandler(t *testing.T) {
	testServer := testServer(t, func(req *collectortracepb.ExportTraceServiceRequest) bool {
		return strings.Contains(req.String(), "alerty panic handler")
		// && strings.Contains(req.String(), "TestRecoverHandler") // TODO: stack trace
	})

	alerty.Start(alerty.AlertyServiceConfig{
		Name:        "test",
		Version:     "1.0.0",
		Environment: "test",
		IngestURL:   testServer.URL,
	})
	defer alerty.Stop()

	doPanicRecover := func() {
		defer alerty.Recover(func(r interface{}) {
			fmt.Println(r)
		})

		panic("alerty panic handler")
	}

	doPanicRecover()
}
