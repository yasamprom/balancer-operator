package model

import (
	"context"
)

type Usecases interface {
	NotifyEvents(ctx context.Context, nodes UpdateNodes) error
}
