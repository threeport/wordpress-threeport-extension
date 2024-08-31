// generated by 'threeport-sdk gen' - do not edit

package util

import (
	echo "github.com/labstack/echo/v4"
	version "github.com/threeport/wordpress-threeport-extension/internal/version"
	"net/http"
)

// RestApiVersion provides the version of the REST API binary.
type RestApiVersion struct {
	Version string `json:"Version" validate:"required"`
}

// VersionRoute adds the /version route to the server to return the API
// server's version.
func VersionRoute(e *echo.Echo) {
	e.GET("/version", func(c echo.Context) error {
		return c.JSON(
			http.StatusOK,
			RestApiVersion{
				Version: version.GetVersion(),
			},
		)
	})
}
