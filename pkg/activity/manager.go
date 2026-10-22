// Copyright (c) 2019-2025 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation

package activity

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redhat-developer/web-terminal-exec/pkg/operations"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

type ActivityManager interface {
	// Start starts tracking users activity and scheduling workspace stopping if there is no activity for idle timeout
	// Should be called once
	Start()

	// Tick registers users activity and postpones workspace stopping by inactivity
	Tick()
}

type noOpManager struct{}

func (*noOpManager) Tick()  {}
func (*noOpManager) Start() {}

type activityManager struct {
	idleTimeout        time.Duration
	stopRetryPeriod    time.Duration
	devworkspaceClient dynamic.Interface
	activityC          chan bool
}

func (m *activityManager) Start() {
	logrus.Infof("DevWorkspace will be stopped automatically in %s if there is no activity", m.idleTimeout)
	timer := time.NewTimer(m.idleTimeout)
	var shutdownChan = make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-timer.C:
				if err := operations.StopDevWorkspace(m.devworkspaceClient); err != nil {
					timer.Reset(m.stopRetryPeriod)
					logrus.Errorf("Failed to stop workspace. Will retry in %s. Cause: %s", m.stopRetryPeriod, err)
				} else {
					logrus.Info("Workspace is successfully stopped by inactivity")
					return
				}
			case <-m.activityC:
				logrus.Debug("Activity is reported. Resetting timer")
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(m.idleTimeout)
			case <-shutdownChan:
				logrus.Info("Received SIGTERM: shutting down activity manager")
				return
			}
		}
	}()
}

func (m *activityManager) Tick() {
	select {
	case m.activityC <- true:
	default:
		// activity is already registered and it will reset timer if workspace won't be stopped
		logrus.Debug("activity manager is temporary busy")
	}
}

func NewActivityManager(idleTimeout, stopRetryPeriod time.Duration, clientProvider operations.ClientProvider) (ActivityManager, error) {
	if idleTimeout < 0 {
		return &noOpManager{}, nil
	}

	if stopRetryPeriod <= 0 {
		return nil, fmt.Errorf("stop retry period must be greater than 0 if idling is enabled")
	}

	devworkspaceClient, _, err := clientProvider.NewDevWorkspaceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes API client: %s", err)
	}
	activityManager := &activityManager{
		idleTimeout:        idleTimeout,
		stopRetryPeriod:    stopRetryPeriod,
		devworkspaceClient: devworkspaceClient,
		activityC:          make(chan bool),
	}
	return activityManager, nil
}
