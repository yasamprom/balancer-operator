package usecases

import (
	"context"

	model "github.com/yasamprom/balancer-operator/internal/model"
)

type SlicerClient interface {
	NotifyEvents(ctx context.Context, nodes model.UpdateNodes) error
}

type usecases struct {
	slicer SlicerClient
}

func NewUsecases(c SlicerClient) *usecases {
	return &usecases{slicer: c}
}
