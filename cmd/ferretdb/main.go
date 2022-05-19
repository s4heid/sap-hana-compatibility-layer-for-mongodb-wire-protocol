// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"

	//"github.com/lucboj/FerretDB_SAP_HANA/internal/pg"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/lucboj/FerretDB_SAP_HANA/internal/clientconn"
	"github.com/lucboj/FerretDB_SAP_HANA/internal/hana"
	"github.com/lucboj/FerretDB_SAP_HANA/internal/handlers"
	"github.com/lucboj/FerretDB_SAP_HANA/internal/util/debug"
	"github.com/lucboj/FerretDB_SAP_HANA/internal/util/logging"
	"github.com/lucboj/FerretDB_SAP_HANA/internal/util/version"
)

//nolint:gochecknoglobals // flags are defined there to be visible in `bin/ferretdb-testcover -h` output
var (
	debugAddrF       = flag.String("debug-addr", "127.0.0.1:8088", "debug address")
	listenAddrF      = flag.String("listen-addr", "127.0.0.1:27017", "listen address")
	modeF            = flag.String("mode", string(clientconn.AllModes[0]), fmt.Sprintf("operation mode: %v", clientconn.AllModes))
	postgresqlURLF   = flag.String("postgresql-url", "postgres://postgres:1234password@127.0.0.1:5433/ferretdb", "PostgreSQL URL")
	proxyAddrF       = flag.String("proxy-addr", "127.0.0.1:37017", "")
	tlsF             = flag.Bool("tls", false, "enable insecure TLS")
	versionF         = flag.Bool("version", false, "print version to stdout (full version, commit, branch, dirty flag) and exit")
	testConnTimeoutF = flag.Duration("test-conn-timeout", 0, "test: set connection timeout")
	saphanaURL       = flag.String("saphana-url", "hdb://User1:Password1@999deec0-ccb7-4a5e-b317-d419e19be648.hana.prod-us10.hanacloud.ondemand.com:443?encrypt=true&sslValidateCertificate=false", "SAP HANA URL")
)

func main() {
	logging.Setup(zap.DebugLevel)
	logger := zap.L()
	flag.Parse()

	info := version.Get()

	if *versionF {
		fmt.Fprintln(os.Stdout, "version:", info.Version)
		fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)
		return
	}

	logger.Info(
		"Starting FerretDB "+info.Version+"...",
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
		zap.String("branch", info.Branch),
		zap.Bool("dirty", info.Dirty),
	)

	var found bool
	for _, m := range clientconn.AllModes {
		if *modeF == string(m) {
			found = true
			break
		}
	}
	if !found {
		logger.Sugar().Fatalf("Unknown mode %q.", *modeF)
	}

	if *tlsF {
		logger.Sugar().Warn("The current TLS implementation is not secure.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, *debugAddrF, logger.Named("debug"))

	//pgPool, err := pg.NewPool(*postgresqlURLF, logger, false)
	hanaPool, err := hana.CreatePool(*saphanaURL, logger, false)
	if err != nil {
		logger.Fatal(err.Error())
	}
	//defer pgPool.Close()
	defer hanaPool.Close()

	listenerMetrics := clientconn.NewListenerMetrics()
	handlersMetrics := handlers.NewMetrics()
	prometheus.DefaultRegisterer.MustRegister(listenerMetrics, handlersMetrics)

	//l := clientconn.NewListener(&clientconn.NewListenerOpts{
	//	ListenAddr:      *listenAddrF,
	//	TLS:             *tlsF,
	//	ProxyAddr:       *proxyAddrF,
	//	Mode:            clientconn.Mode(*modeF),
	//	PgPool:          pgPool,
	//	Logger:          logger.Named("listener"),
	//	Metrics:         listenerMetrics,
	//	HandlersMetrics: handlersMetrics,
	//	TestConnTimeout: *testConnTimeoutF,
	//})

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      *listenAddrF,
		TLS:             *tlsF,
		ProxyAddr:       *proxyAddrF,
		Mode:            clientconn.Mode(*modeF),
		HanaPool:        hanaPool,
		Logger:          logger.Named("listener"),
		Metrics:         listenerMetrics,
		HandlersMetrics: handlersMetrics,
		TestConnTimeout: *testConnTimeoutF,
	})

	err = l.Run(ctx)
	if err == nil || err == context.Canceled {
		logger.Info("Listener stopped")
	} else {
		logger.Error("Listener stopped", zap.Error(err))
	}

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic(err)
	}
	for _, mf := range mfs {
		if _, err := expfmt.MetricFamilyToText(os.Stderr, mf); err != nil {
			panic(err)
		}
	}
}
