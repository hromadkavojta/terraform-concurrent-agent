package agent

import (
	"github.com/labstack/echo"
	"go/types"
	"net/http"
)

type Service types.Struct

func NewService() *Service {
	return &Service{}
}

func (s *Service) TerraformIt(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}
