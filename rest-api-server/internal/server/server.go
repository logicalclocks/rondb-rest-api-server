/*
 * This file is part of the RonDB REST API Server
 * Copyright (c) 2022 Hopsworks AB
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
	"hopsworks.ai/rdrs/internal/handlers"
	"hopsworks.ai/rdrs/internal/log"
	"hopsworks.ai/rdrs/internal/security/apikey"
	"hopsworks.ai/rdrs/internal/security/tlsutils"
	"hopsworks.ai/rdrs/internal/server/grpcsrv"
	"hopsworks.ai/rdrs/pkg/api"
	"hopsworks.ai/rdrs/version"
	// _ "github.com/ianlancetaylor/cgosymbolizer" // enable this for stack trace of c layer
)

type Router interface {
	SetupRouter(registerHandlers []handlers.RegisterHandlers) error
	StartRouter() error
	StopRouter() error
	GetServer() (*http.Server, *grpc.Server)
}

type RouterConext struct {
	// REST Server
	RESTServerIP   string
	RESTServerPort uint16
	GRPCServerIP   string
	GRPCServerPort uint16
	APIVersion     string
	Engine         *gin.Engine

	// RonDB
	DBIP   string
	DBPort uint16

	//server
	HttpServer *http.Server
	GRPCServer *grpc.Server
}

var _ Router = (*RouterConext)(nil)

func (rc *RouterConext) SetupRouter(handlers []handlers.RegisterHandlers) error {
	gin.SetMode(gin.ReleaseMode)
	rc.Engine = gin.New()

	for _, handler := range handlers {
		handler(rc.Engine)
	}

	// connect to RonDB
	dal.InitializeBuffers()
	err := dal.InitRonDBConnection(fmt.Sprintf("%s:%d", rc.DBIP, rc.DBPort), true)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", rc.RESTServerIP, rc.RESTServerPort)
	rc.HttpServer = &http.Server{
		Addr:    address,
		Handler: rc.Engine,
	}

	return nil
}

func (rc *RouterConext) StartRouter() error {

	log.Infof("REST Server Listening on %s:%d, GRPC Server Listening on %s:%d ",
		rc.RESTServerIP, rc.RESTServerPort, rc.GRPCServerIP, rc.GRPCServerPort)

	var serverTLS *tls.Config
	var err error

	if config.Configuration().Security.EnableTLS {
		if config.Configuration().Security.CertificateFile == "" ||
			config.Configuration().Security.PrivateKeyFile == "" {
			return fmt.Errorf("Server Certificate/Key not set")
		}

		serverTLS, err = serverTLSConfig()
		if err != nil {
			return fmt.Errorf("Unable to set server TLS config. Error %v", err)
		}
	}

	go func() { // Start REST Server

		if config.Configuration().Security.EnableTLS {
			rc.HttpServer.TLSConfig = serverTLS
			err = rc.HttpServer.ListenAndServeTLS(config.Configuration().Security.CertificateFile,
				config.Configuration().Security.PrivateKeyFile)
		} else {
			err = rc.HttpServer.ListenAndServe()
		}
		if err != nil {
			log.Infof("Http server returned. Error: %v", err)
		}
	}()

	go func() { // Start GRPC Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", rc.GRPCServerIP, rc.GRPCServerPort))
		if err != nil {
			log.Fatalf("GRPC server returned. Error: %v", err)
		}
		rc.GRPCServer = grpc.NewServer()
		GRPCServer := grpcsrv.GetGRPCServer()
		api.RegisterRonDBRESTServer(rc.GRPCServer, GRPCServer)
		rc.GRPCServer.Serve(lis)
	}()

	return nil
}

func serverTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
	}

	if config.Configuration().Security.RequireAndVerifyClientCert {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if config.Configuration().Security.RootCACertFile != "" {
		tlsConfig.ClientCAs = tlsutils.TrustedCAs(config.Configuration().Security.RootCACertFile)
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func (rc *RouterConext) StopRouter() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop REST Server
	if err := rc.HttpServer.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	// Stop GRPC Server
	rc.GRPCServer.Stop()

	// Stop RonDB Connection
	dalErr := dal.ShutdownConnection()
	dal.ReleaseAllBuffers()

	if dalErr != nil {
		log.Errorf("Failed to stop RonDB API. Error %v", dalErr)
	}

	// Clean API Key Cache
	apikey.Reset()

	return nil
}

func CreateRouterContext() Router {
	router := RouterConext{
		RESTServerIP:   config.Configuration().RestServer.RESTServerIP,
		RESTServerPort: config.Configuration().RestServer.RESTServerPort,

		GRPCServerIP:   config.Configuration().RestServer.GRPCServerIP,
		GRPCServerPort: config.Configuration().RestServer.GRPCServerPort,

		APIVersion: version.API_VERSION,

		DBIP:   config.Configuration().RonDBConfig.IP,
		DBPort: config.Configuration().RonDBConfig.Port,

		HttpServer: &http.Server{},
	}
	return &router
}

func (rc *RouterConext) GetServer() (*http.Server, *grpc.Server) {
	return rc.HttpServer, rc.GRPCServer
}
