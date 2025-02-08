package api

import (
	"net/http"

	"receipt_processor/pkg/middleware"
	"receipt_processor/pkg/service"
)

// Router is the API router that ties the HTTP endpoints to the service layer.
type Router struct {
	receiptService service.IReceiptService
	middlewares    []middleware.Middleware
}

// NewRouter creates a new HTTP handler with the defined routes and applies the given middleware.
func NewRouter(rs service.IReceiptService, mws []middleware.Middleware) http.Handler {
	r := &Router{
		receiptService: rs,
		middlewares:    mws,
	}

	// Using the standard ServeMux.
	mux := http.NewServeMux()
	// Register the process receipt endpoint.
	mux.Handle("/receipts/process", applyMiddlewares(http.HandlerFunc(r.ProcessReceiptHandler), mws))
	// Register the get points endpoint. Since the route includes a dynamic receipt ID,
	// we register a prefix and then parse the ID within the handler.
	mux.Handle("/receipts/", applyMiddlewares(http.HandlerFunc(r.GetPointsHandler), mws))

	return mux
}

// applyMiddlewares composes the middleware functions around the handler.
func applyMiddlewares(h http.Handler, mws []middleware.Middleware) http.Handler {
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}
