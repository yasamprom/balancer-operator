package usecases

import (
	"context"
	"log"

	model "github.com/yasamprom/balancer-operator/internal/model"
)

func (uc usecases) NotifyEvents(ctx context.Context, nodes model.UpdateNodes) error {
	log.Println("Notify events...")
	return uc.slicer.NotifyState(ctx, nodes)
}
