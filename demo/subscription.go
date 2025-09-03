package demo

import (
	"time"

	"github.com/pritunl/pritunl-cloud/subscription"
)

var Subscription = &subscription.Subscription{
	Active:            true,
	Status:            "active",
	Plan:              "zero",
	Quantity:          1,
	Amount:            5000,
	PeriodEnd:         time.Unix(1893499200, 0),
	TrialEnd:          time.Time{},
	CancelAtPeriodEnd: false,
	Balance:           0,
	UrlKey:            "demo",
}
