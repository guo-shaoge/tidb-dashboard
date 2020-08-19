// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package queryeditor

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pingcap-incubator/tidb-dashboard/pkg/apiserver/user"
	"github.com/pingcap-incubator/tidb-dashboard/pkg/apiserver/utils"
	"github.com/pingcap-incubator/tidb-dashboard/pkg/config"
	"github.com/pingcap-incubator/tidb-dashboard/pkg/tidb"
)

type ServiceParams struct {
	fx.In
	Config     *config.Config
	TiDBClient *tidb.Client
}

type Service struct {
	params       ServiceParams
	lifecycleCtx context.Context
}

func NewService(lc fx.Lifecycle, p ServiceParams) *Service {
	service := &Service{params: p}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			service.lifecycleCtx = ctx
			return nil
		},
	})

	return service
}

func Register(r *gin.RouterGroup, auth *user.AuthService, s *Service) {
	endpoint := r.Group("/query_editor")
	endpoint.Use(auth.MWAuthRequired())
	endpoint.Use(utils.MWConnectTiDB(s.params.TiDBClient))
	endpoint.Use(utils.MWForbidByExperimentalFlag(s.params.Config.EnableExperimental))
	endpoint.POST("/run", s.runHandler)
}

type RunRequest struct {
	Statements string `json:"statements" example:"show databases;"`
	MaxRows    int    `json:"max_rows" example:"1000"`
}

type RunResponse struct {
	ErrorMsg    string          `json:"error_msg"`
	ColumnNames []string        `json:"column_names"`
	Rows        [][]interface{} `json:"rows"`
	ExecutionMs int64           `json:"execution_ms"`
	ActualRows  int             `json:"actual_rows"`
}

func executeStatements(context context.Context, db *sql.DB, statements string) ([]string, [][]interface{}, error) {
	rows, err := db.QueryContext(context, statements)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	retRows := make([][]interface{}, 0)

	values := make([]sql.RawBytes, len(colNames))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, nil, err
		}

		retRow := make([]interface{}, 0, len(values))
		var value interface{}
		for _, col := range values {
			if col == nil {
				value = nil
			} else {
				value = string(col)
			}
			retRow = append(retRow, value)
		}
		retRows = append(retRows, retRow)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return colNames, retRows, nil
}

// @ID queryEditorRun
// @Summary Run
// @Description Run statements
// @Produce json
// @Param request body RunRequest true "Request body"
// @Success 200 {object} RunResponse
// @Router /query_editor/run [post]
// @Security JwtAuth
// @Failure 400 {object} utils.APIError "Bad request"
// @Failure 401 {object} utils.APIError "Unauthorized failure"
// @Failure 403 {object} utils.APIError "Experimental feature not enabled"
func (s *Service) runHandler(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	ctx, cancel := context.WithTimeout(s.lifecycleCtx, time.Minute*5)
	defer cancel()

	startTime := time.Now()
	colNames, rows, err := executeStatements(ctx, utils.GetTiDBConnection(c).DB(), req.Statements)
	elapsedTime := time.Since(startTime)

	if err != nil {
		log.Warn("Failed to execute user input statements", zap.String("statements", req.Statements), zap.Error(err))
		c.JSON(http.StatusOK, RunResponse{
			ErrorMsg:    err.Error(),
			ColumnNames: nil,
			Rows:        nil,
			ExecutionMs: elapsedTime.Milliseconds(),
			ActualRows:  0,
		})
		return
	}

	truncatedRows := rows
	if len(truncatedRows) > req.MaxRows {
		truncatedRows = truncatedRows[:req.MaxRows]
	}

	c.JSON(http.StatusOK, RunResponse{
		ColumnNames: colNames,
		Rows:        truncatedRows,
		ExecutionMs: elapsedTime.Milliseconds(),
		ActualRows:  len(rows),
	})
}
