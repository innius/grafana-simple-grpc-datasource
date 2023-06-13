package framer

import (
	"testing"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/stretchr/testify/assert"
)

func TestFormatDisplayName(t *testing.T) {
	t.Run("DisplayNameFormatter", func(t *testing.T) {
		t.Run("should return an empty string if display name is not specified", func(t *testing.T) {
			res := formatDisplayName(FormatDisplayNameInput{MetricID: "foo"})
			assert.Empty(t, res)
		})
		t.Run("should return an error empty string", func(t *testing.T) {
			res := formatDisplayName(FormatDisplayNameInput{DisplayName: "{{.Foo}}"})
			assert.Empty(t, res)
		})
		t.Run("should return a formatted display name", func(t *testing.T) {
			input := FormatDisplayNameInput{
				DisplayName: ">>{{metric}}-{{dim}}-{{label}}-{{arg}}<<",
				MetricID:    "a",
				Dimensions: []models.Dimension{
					{
						Key:   "dim",
						Value: "b",
					},
				},
				Labels: []*pb.Label{
					{
						Key:   "label",
						Value: "c",
					},
				},
				Args: []Arg{{
					Key:   "arg",
					Value: "d",
				}},
			}

			res := formatDisplayName(input)
			assert.Equal(t, ">>a-b-c-d<<", res)
		})
	})
}
