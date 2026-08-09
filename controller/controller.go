package controller

import (
	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/scheduler/abstraction"
)

type Controller struct {
	scheduler abstraction.Ischeduler
	socket    string
}

func NewController(config context.ControllerConfig, s abstraction.Ischeduler) *Controller {
	return &Controller{
		scheduler: s,
		socket:    config.Socket,
	}
}
