package controller

import (
	"github.com/stevenchiu30801/bans5gc-operator/pkg/controller/bansslice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, bansslice.Add)
}
