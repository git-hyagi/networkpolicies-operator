package controller

import (
	"github.com/lab/networkpolicies-operator/pkg/controller/forcenetpol"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, forcenetpol.Add)
}
