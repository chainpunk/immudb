/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tc

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rs/cors"
)

func (s *ImmuTcServer) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cliOpts := &immuclient.Options{
		Address:            s.Options.ImmudbAddress,
		Port:               s.Options.ImmudbPort,
		HealthCheckRetries: 1,
		MTLs:               s.Options.MTLs,
		MTLsOptions:        s.Options.MTLsOptions,
		Auth:               false,
		Config:             "",
	}

	ic, err := immuclient.NewImmuClient(cliOpts)
	if err != nil {
		s.Logger.Errorf("Unable to instantiate client %s", err)
		return err
	}

	c := NewImmuTc(ic)
	c.Start(ctx)
	mux := runtime.NewServeMux()

	handler := cors.Default().Handler(mux)

	s.installShutdownHandler()
	s.Logger.Infof("starting immutc: %v", s.Options)
	if s.Options.Pidfile != "" {
		if s.Pid, err = server.NewPid(s.Options.Pidfile); err != nil {
			s.Logger.Errorf("failed to write pidfile: %s", err)
			return err
		}
	}

	go func() {
		if err = http.ListenAndServe(s.Options.Address+":"+strconv.Itoa(s.Options.Port), handler); err != nil && err != http.ErrServerClosed {
			s.Logger.Errorf("Unable to launch immutc:%+s\n", err)
		}
	}()
	<-s.quit
	return err
}

func (s *ImmuTcServer) Stop() error {
	s.Logger.Infof("stopping immutc: %v", s.Options)
	defer func() { s.quit <- struct{}{} }()
	return nil
}

func (s *ImmuTcServer) installShutdownHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer func() {
			s.quit <- struct{}{}
		}()
		<-c
		s.Logger.Infof("caught SIGTERM")
		if err := s.Stop(); err != nil {
			s.Logger.Errorf("shutdown error: %v", err)
		}
		s.Logger.Infof("shutdown completed")
	}()
}
