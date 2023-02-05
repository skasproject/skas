package tokenstore

import (
	"context"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var cleanerLog = ctrl.Log.WithName("Cleaner")

type Cleaner struct {
	Period     time.Duration
	TokenStore TokenStore
}

func (*Cleaner) NeedLeaderElection() bool {
	return false
}

func (c *Cleaner) Run(ctx context.Context) error {
	return c.Start(ctx)
}

func (c *Cleaner) Start(ctx context.Context) error {
	if c.Period == 0 {
		c.Period = 30 * time.Second
	}
	cleanerLog.Info("Cleaner start")
	go wait.Until(func() {
		_ = c.TokenStore.Clean() // Error has been logged by Clean()
	}, c.Period, ctx.Done())
	<-ctx.Done()
	cleanerLog.Info("Cleaner shutdown")
	return nil
}
