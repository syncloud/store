package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net"
	"os"
)

type IndexRefresher interface {
	Refresh() error
}

type Api struct {
	index IndexRefresher
	echo  *echo.Echo
}

func NewApi(index IndexRefresher) *Api {
	return &Api{
		echo:  echo.New(),
		index: index,
	}
}

func (a *Api) Start() error {
	a.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	a.echo.Use(middleware.Recover())

	a.echo.POST("/refresh", a.Refresh)

	_ = os.RemoveAll(InternalApi)
	l, err := net.Listen("unix", InternalApi)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	a.echo.Listener = l
	go func() {
		a.echo.Start("")
	}()
	return nil
}

func (a *Api) Refresh(_ echo.Context) error {
	return a.index.Refresh()
}
