package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/model"
)

func registerPublishRoutes(e *echo.Echo, binary *SnapBinaryPublisher, yaml *SnapYamlPublisher, icon *IconPublisher) {
	e.POST("/syncloud/v1/publish/snap/init", func(c echo.Context) error {
		var req model.PublishInitRequest
		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		resp, err := binary.Init(req)
		return reply(c, resp, err)
	})
	e.POST("/syncloud/v1/publish/snap/part-url", func(c echo.Context) error {
		var req model.PublishPartUrlRequest
		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		resp, err := binary.PartUrl(req)
		return reply(c, resp, err)
	})
	e.POST("/syncloud/v1/publish/snap/finalise", func(c echo.Context) error {
		var req model.PublishFinaliseRequest
		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		resp, err := binary.Finalise(req)
		return reply(c, resp, err)
	})
	e.POST("/syncloud/v1/publish/snap-yaml", func(c echo.Context) error {
		var req model.PublishSnapYamlRequest
		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		resp, err := yaml.Publish(req)
		return reply(c, resp, err)
	})
	e.POST("/syncloud/v1/publish/icon", func(c echo.Context) error {
		var req model.PublishIconRequest
		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		resp, err := icon.Publish(req)
		return reply(c, resp, err)
	})
}

func reply(c echo.Context, resp interface{}, err error) error {
	if err == nil {
		return c.JSON(http.StatusOK, resp)
	}
	var ae *apiError
	if errors.As(err, &ae) {
		return c.String(ae.Status, ae.Msg)
	}
	return c.String(http.StatusInternalServerError, err.Error())
}
