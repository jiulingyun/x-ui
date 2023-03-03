package controller

import (
	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type ApiController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
}

func NewApiController(g *gin.RouterGroup) *ApiController {
	a := &ApiController{}
	a.initRouter(g)
	return a
}

func (a *ApiController) initRouter(g *gin.RouterGroup) {
	g.GET("/api/get_sys_status", a.getSysStatus)
}

func (a *ApiController) getSysStatus(c *gin.Context) {
	inbounds, err := a.inboundService.GetInbounds(1)
	if err != nil {
		jsonMsg(c, "获取失败", err)
		return
	}
	jsonObj(c, inbounds, nil)
}
