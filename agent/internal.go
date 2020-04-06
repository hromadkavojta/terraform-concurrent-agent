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

//TODO this should be called on github webook ( make terraform plan and save it )
func (s *Service) TerraformPlan(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}

//TODO this should send client info about new terraform plan ( infrastructure config)
func (s *Service) TerraformShow(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}

//TODO this should apply terraform config from given plan
func (s *Service) TerraformApply(c echo.Context) error {

	return c.JSON(http.StatusOK, nil)
}
