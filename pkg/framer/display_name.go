package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"bytes"
	"strings"
	"text/template"
)

type FormatDisplayNameInput struct {
	DisplayName string
	MetricID    string
	FieldName   string
	Dimensions  []models.Dimension
	Labels      []*pb.Label
	Args        []Arg
}

func (f FormatDisplayNameInput) ToMap() map[string]string {
	dict := map[string]string{
		"metric": f.MetricID,
		"field":  f.FieldName,
	}
	for _, v := range f.Dimensions {
		dict[v.Key] = v.Value
	}
	for i := range f.Labels {
		label := f.Labels[i]
		dict[label.Key] = label.Value
	}
	for i := range f.Args {
		arg := f.Args[i]
		dict[arg.Key] = arg.Value
	}
	return dict
}

func formatDisplayName(input FormatDisplayNameInput) string {
	if input.DisplayName == "" {
		return ""
	}
	s, err := parseDisplayNameExpr(input.ToMap(), input.DisplayName)
	if err != nil {
		return ""
	}
	return s
}

type Arg struct {
	Key   string
	Value string
}

func parseTemplate(alias string) (*template.Template, error) {
	t := template.New("alias")
	text := strings.Replace(alias, "{{", "{{.", -1)
	return t.Parse(text)
}

func parseDisplayNameExpr(ctx map[string]string, alias string) (string, error) {
	t, err := parseTemplate(alias)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, ctx)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
