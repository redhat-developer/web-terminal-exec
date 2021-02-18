//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package main

import (
	"net/http"

	"github.com/eclipse/che-machine-exec/activity"
	"github.com/eclipse/che-machine-exec/auth"
	commonRest "github.com/eclipse/che-machine-exec/common/rest"

	jsonrpc "github.com/eclipse/che-go-jsonrpc"
	jsonRpcApi "github.com/eclipse/che-machine-exec/api/jsonrpc"
	"github.com/eclipse/che-machine-exec/api/rest"
	"github.com/eclipse/che-machine-exec/api/websocket"
	"github.com/eclipse/che-machine-exec/cfg"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg.Parse()
	cfg.Print()

	var activityManager activity.Manager = nil

	r := gin.Default()

	if cfg.StaticPath != "" {
		r.StaticFS("/static", http.Dir(cfg.StaticPath))
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/static")
		})
	}

	// connect to exec api endpoint(websocket with json-rpc)
	r.GET("/connect", func(c *gin.Context) {
		websocket.HandleConnect(c)
	})

	// attach to get exec output and sent user input(by simple websocket)
	// Todo: rework to use only one websocket connection https://github.com/eclipse/che-machine-exec/issues/4
	r.GET("/attach/:id", func(c *gin.Context) {
		websocket.HandleAttach(c)
	})

	r.POST("/exec/config", func(c *gin.Context) {
		rest.HandleKubeConfig(c)
	})

	r.POST("/exec/init", func(c *gin.Context) {
		rest.HandleInit(c)
	})

	r.POST("/activity/tick", func(c *gin.Context) {
		if activityManager == nil {
			activityManager = initializeActivityManager(c)

			if activityManager != nil {
				rest.HandleActivityTick(c, activityManager)
			}
		} else {
			rest.HandleActivityTick(c, activityManager)
		}
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
	})

	// create json-rpc routs group
	appOpRoutes := []jsonrpc.RoutesGroup{
		jsonRpcApi.RPCRoutes,
	}
	// register routes
	jsonrpc.RegRoutesGroups(appOpRoutes)
	jsonrpc.PrintRoutes(appOpRoutes)

	if cfg.UseTLS {
		if err := r.RunTLS(cfg.URL, "/var/serving-cert/tls.crt", "/var/serving-cert/tls.key"); err != nil {
			logrus.Fatal("Unable to start server with TLS enabled. Cause: ", err.Error())
		}
	} else {
		if err := r.Run(cfg.URL); err != nil {
			logrus.Fatal("Unable to start server. Cause: ", err.Error())
		}
	}
}

func initializeActivityManager(c *gin.Context) activity.Manager {
	var token string
	if auth.IsEnabled() {
		var err error
		token, err = auth.Authenticate(c)
		if err != nil {
			commonRest.WriteErrorResponse(c, err)
			return nil
		}
	}
	activityManager, err := activity.New(cfg.IdleTimeout, cfg.StopRetryPeriod, token)
	if err != nil {
		logrus.Fatal("Unable to create activity manager. Cause: ", err.Error())
		return nil
	}

	activityManager.Start()
	return activityManager
}
