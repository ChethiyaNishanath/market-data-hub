package ws

import (
	"context"

	"github.com/ChethiyaNishanath/market-data-hub/internal/events"
)

type Module struct {
	Handler *Handler
	Router  *Router
	ConnMgr *ConnectionManager
}

func NewModule(ctx *context.Context, connMgr *ConnectionManager, bus *events.Bus) *Module {
	router := NewRouter()

	handler := NewHandler(router, connMgr)

	router.Handle("subscribe", handler.HandleSubscribe)
	router.Handle("unsubscribe", handler.HandleUnsubscribe)

	return &Module{
		Handler: handler,
		Router:  router,
		ConnMgr: connMgr,
	}
}
