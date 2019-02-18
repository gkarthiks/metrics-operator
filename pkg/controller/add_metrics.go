package controller

import (
	"github.com/gkarthiks/metrics-operator/pkg/controller/metrics"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, metrics.Add)
}
