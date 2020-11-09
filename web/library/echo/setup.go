package setup

import (
	"github.com/Seven4X/link/web/library/config"
	"github.com/Seven4X/link/web/library/echo/validator"
	"github.com/Seven4X/link/web/library/log"
	adapter "github.com/alibaba/sentinel-golang/adapter/echo"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/nacos"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEcho() (e *echo.Echo) {
	// Echo instance
	e = echo.New()
	e.Validator = validator.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//sentinel参考：https://github.com/alibaba/sentinel-golang/tree/master/adapter/echo
	//https://github.com/alibaba/sentinel-golang/blob/master/example/datasource/nacos/datasource_nacos_example.go

	initSentinel()
	//全局限流
	e.Use(adapter.SentinelMiddleware())
	//ip限流
	e.Use(
		adapter.SentinelMiddleware(
			// customize resource extractor if required
			// method_path by default
			adapter.WithResourceExtractor(func(ctx echo.Context) string {
				if res, ok := ctx.Get("X-Real-IP").(string); ok {
					return res
				}
				return ""
			}),
			// customize block fallback if required
			// abort with status 429 by default
			adapter.WithBlockFallback(func(ctx echo.Context) error {
				return ctx.JSON(400, map[string]interface{}{
					"err":  "too many requests; the quota used up",
					"code": 10222,
				})
			}),
		),
	)
	return e
}

func initSentinel() {
	err := sentinel.InitDefault()
	if err != nil {
		// 初始化 Sentinel 失败
		log.Error("initSentinel-error", err.Error())
	}
	//从acm加载配置
	//rule配置参考flow.rule
	h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
	client := config.GetAcmClient()
	nds, err := nacos.NewNacosDataSource(client, "link-hub-go", "flow", h)
	if err != nil {
		log.Warnf("Fail to create nacos data source client, err: %+v", err)
		return
	}
	err = nds.Initialize()
	if err != nil {
		log.Warnf("Fail to initialize nacos data source client, err: %+v", err)
		return
	}
}