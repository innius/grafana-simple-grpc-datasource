package backendapi

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func backendErrorResponse(err error) (data.Frames, error) {
	st := status.Convert(err)
	backend.Logger.Error(st.Code().String(), "error", err)
	return nil, convertBackendError(st)
}

// convertBackendError converts a grpc status code to a plugin error message
func convertBackendError(st *status.Status) error {
	switch st.Code() {
	case codes.Unauthenticated:
		return errors.New("Authentication error; please check if your datasource is provided with valid credentials")
	case codes.Unavailable:
		return errors.New("Could not establish a connection; please check if your datasource is provided with valid credentials")
	case codes.DeadlineExceeded:
		return errors.New("Query did not complete within the expected timeframe; please check your query configuration or try to select a smaller period")
	case codes.ResourceExhausted:
		return errors.New("Too many concurrent queries; reduce the number of concurrent queries to circumvent this issue")
	case codes.NotFound:
		return errors.New("One or more resources cannot be found; please check your query configuration")
	case codes.FailedPrecondition:
		return errors.New("Query cannot be processed because of invalid configuration; please check your query configuration")
	default:
		return errors.New("Query cannot be processed because of an internal error")
	}
}
