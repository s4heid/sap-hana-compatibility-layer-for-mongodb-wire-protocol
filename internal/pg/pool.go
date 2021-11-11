// Copyright 2021 Baltoro OÜ.
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

package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(connString string, logger *zap.Logger, lazy bool) (*Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("pg.NewPool: %w", err)
	}

	config.LazyConnect = lazy

	// That only affects text protocol; pgx mostly uses a binary one.
	// See:
	// * https://github.com/jackc/pgx/issues/520
	// * https://github.com/jackc/pgx/issues/789
	// * https://github.com/jackc/pgx/issues/863
	// * https://github.com/MangoDB-io/MangoDB/issues/43
	config.ConnConfig.RuntimeParams["timezone"] = "UTC"

	config.ConnConfig.RuntimeParams["application_name"] = "MangoDB"
	config.ConnConfig.RuntimeParams["search_path"] = ""

	if logger.Core().Enabled(zap.DebugLevel) {
		config.ConnConfig.LogLevel = pgx.LogLevelTrace
		config.ConnConfig.Logger = zapadapter.NewLogger(logger.Named("pgconn.Pool"))
	}

	p, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("pg.NewPool: %w", err)
	}

	if !lazy {
		rows, err := p.Query(context.Background(), "SHOW ALL")
		if err != nil {
			return nil, fmt.Errorf("pg.NewPool: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name, setting, description string
			if err := rows.Scan(&name, &setting, &description); err != nil {
				return nil, fmt.Errorf("pg.NewPool: %w", err)
			}

			switch name {
			case "server_encoding":
				if setting != "UTF8" {
					return nil, fmt.Errorf("pg.NewPool: %q is %q, want %q", name, setting, "UTF8")
				}
			case "client_encoding":
				if setting != "UTF8" {
					return nil, fmt.Errorf("pg.NewPool: %q is %q, want %q", name, setting, "UTF8")
				}
			case "lc_collate":
				if setting != "C" && setting != "POSIX" && setting != "en_US.utf8" {
					return nil, fmt.Errorf("pg.NewPool: %q is %q", name, setting)
				}
			case "lc_ctype":
				if setting != "C" && setting != "POSIX" && setting != "en_US.utf8" {
					return nil, fmt.Errorf("pg.NewPool: %q is %q", name, setting)
				}
			default:
				continue
			}

			logger.Debug("PostgreSQL setting", zap.String("name", name), zap.String("setting", setting))
		}
	}

	return &Pool{
		Pool: p,
	}, nil
}
