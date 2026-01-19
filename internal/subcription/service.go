package subcription

import (
	"github.com/ChethiyaNishanath/market-data-hub/internal/domain/subscription"
	"github.com/ChethiyaNishanath/market-data-hub/internal/exchange/binance"
)

type Service struct {
	Handler *Handler
	Router  *binance.Router
	ConnMgr subscription.ClientConnectionManager
}

func NewService(connMgr subscription.ClientConnectionManager) *Service {

	router := binance.NewRouter()
	handler := NewHandler(router, connMgr)

	router.Handle(Subscribe, handler.HandleSubscribe)
	router.Handle(Unsubscribe, handler.HandleUnsubscribe)

	return &Service{
		Handler: handler,
		Router:  router,
		ConnMgr: connMgr,
	}
}
