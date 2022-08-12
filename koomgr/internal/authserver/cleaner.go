/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/

package authserver

import (
	"context"
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

func (this *Cleaner) Start(ctx context.Context) error {
	if this.Period == 0 {
		this.Period = 30 * time.Second
	}
	cleanerlog.Info("Cleaner start")
	go wait.Until(func() {
		this.TokenBasket.Clean()
	}, this.Period, ctx.Done())
	<-ctx.Done()
	cleanerlog.Info("Cleaner shutdown")
	return nil
}
