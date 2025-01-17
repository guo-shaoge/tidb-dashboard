// Copyright 2023 PingCAP, Inc. Licensed under Apache-2.0.

package endpoint

import (
	"github.com/joomcode/errorx"
)

var (
	ErrNS               = errorx.NewNamespace("debug_api.endpoint")
	ErrUnknownComponent = ErrNS.NewType("unknown_component")
)
