package controller

import (
	"operator/mysql/pkg/controller/mysql"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, mysql.Add)
}
