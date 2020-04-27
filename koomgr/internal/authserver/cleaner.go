package authserver

import (
	"github.com/koobind/koobind/koomgr/internal/token"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var cleanerlog = ctrl.Log.WithName("Cleaner")

type Cleaner struct {
	Period      time.Duration
	TokenBasket token.TokenBasket
}

func (*Cleaner) NeedLeaderElection() bool {
	return false
}

func (this *Cleaner) Start(stop <-chan struct{}) error {
	if this.Period == 0 {
		this.Period = 30 * time.Second
	}
	cleanerlog.Info("Cleaner start")
	go wait.Until(func() {
		this.TokenBasket.Clean()
	}, this.Period, stop)
	<-stop
	cleanerlog.Info("Cleaner shutdown")
	return nil
}
