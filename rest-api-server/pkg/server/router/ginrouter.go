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

package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hopsworks.ai/rdrs/internal/config"
	"hopsworks.ai/rdrs/internal/dal"
	"hopsworks.ai/rdrs/internal/log"
	"hopsworks.ai/rdrs/internal/router/handler"
	// _ "github.com/ianlancetaylor/cgosymbolizer" // enable this for stack trace for c layer
)

type RouterConext struct {
	// REST Server
	ServerIP   string
	ServerPort uint16
	APIVersion string
	Engine     *gin.Engine

	// RonDB
	DBIP   string
	DBPort uint16

	//server
	Server *http.Server
}

var _ Router = (*RouterConext)(nil)

func (rc *RouterConext) SetupRouter(handlers []handler.RegisterTestHandler) error {
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

	address := fmt.Sprintf("%s:%d", rc.ServerIP, rc.ServerPort)
	rc.Server = &http.Server{
		Addr:    address,
		Handler: rc.Engine,
	}

	return nil
}

func (rc *RouterConext) StartRouter() error {

	log.Infof("Listening on %s", rc.Server.Addr)

	go func() {
		err := rc.Server.ListenAndServe()
		if err != nil {
			log.Infof("Server returned. Error: %v", err)
		}
	}()

	return nil
}

func (rc *RouterConext) StopRouter() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rc.Server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	dalErr := dal.ShutdownConnection()
	dal.ReleaseAllBuffers()

	if dalErr != nil {
		log.Errorf("Failed to stop RonDB API. Error %v", dalErr)
	}

	return nil
}

func CreateRouterContext() Router {
	router := RouterConext{
		ServerIP:   config.Configuration().RestServer.IP,
		ServerPort: config.Configuration().RestServer.Port,
		APIVersion: config.Configuration().RestServer.APIVersion,
		DBIP:       config.Configuration().RonDBConfig.IP,
		DBPort:     config.Configuration().RonDBConfig.Port,
		Server:     &http.Server{},
	}
	return &router
}

func (rc *RouterConext) GetServer() *http.Server {
	return rc.Server
}
