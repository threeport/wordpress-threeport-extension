// generated by 'threeport-sdk gen' - do not edit

package routes

import (
	echo "github.com/labstack/echo/v4"
	handlers "github.com/threeport/wordpress-threeport-extension/pkg/api-server/v0/handlers"
)

// AddRoutes adds routes for all objects of a particular API version.
func AddRoutes(e *echo.Echo, h *handlers.Handler) {
	WordpressDefinitionRoutes(e, h)
	WordpressInstanceRoutes(e, h)
}