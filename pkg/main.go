package main

import (
	"os"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/server"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {

	if err := datasource.Manage("simple-grpc-datasource", server.NewServerInstance, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
