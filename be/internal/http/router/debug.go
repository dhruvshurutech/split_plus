package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type routeInfo struct {
	Method string
	Route  string
}

func PrintRoutes(r chi.Router) {
	var routes []routeInfo

	_ = chi.Walk(r, func(
		method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler,
	) error {
		if method != "" {
			routes = append(routes, routeInfo{
				Method: method,
				Route:  route,
			})
		}
		return nil
	})

	if len(routes) == 0 {
		fmt.Println("No routes found")
		return
	}

	// Calculate column widths
	maxMethodLen := len("METHOD")
	maxRouteLen := len("ROUTE")
	for _, r := range routes {
		if len(r.Method) > maxMethodLen {
			maxMethodLen = len(r.Method)
		}
		if len(r.Route) > maxRouteLen {
			maxRouteLen = len(r.Route)
		}
	}

	// Add padding
	maxMethodLen += 2
	maxRouteLen += 2

	// Print header with borders
	topBorder := "┌" + strings.Repeat("─", maxMethodLen+2) + "┬" + strings.Repeat("─", maxRouteLen+2) + "┐"
	header := fmt.Sprintf("│ %-*s │ %-*s │", maxMethodLen, "METHOD", maxRouteLen, "ROUTE")
	separator := "├" + strings.Repeat("─", maxMethodLen+2) + "┼" + strings.Repeat("─", maxRouteLen+2) + "┤"
	bottomBorder := "└" + strings.Repeat("─", maxMethodLen+2) + "┴" + strings.Repeat("─", maxRouteLen+2) + "┘"

	fmt.Println("\n" + topBorder)
	fmt.Println(header)
	fmt.Println(separator)

	// Print routes
	for _, r := range routes {
		fmt.Printf("│ %-*s │ %-*s │\n", maxMethodLen, r.Method, maxRouteLen, r.Route)
	}

	// Print footer
	fmt.Println(bottomBorder)
	fmt.Printf("\nTotal: %d routes\n", len(routes))
}
