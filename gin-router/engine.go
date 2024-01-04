package ginrouter

import "github.com/gin-gonic/gin"

type decoratorEngine struct {
	decoratorGroup

	engine *gin.Engine
}

var _ iDecoratorRouter = (*decoratorEngine)(nil)

func New(engine *gin.Engine) *decoratorEngine {
	dEngine := &decoratorEngine{
		decoratorGroup: decoratorGroup{
			root:          true,
			filters:       make([]*filter, 0, 16),
			unusedFilters: make([]*filter, 0, 8),
			routerGroup:   &engine.RouterGroup,
		},
		engine: engine,
	}
	dEngine.decoratorGroup.decoratorEngine = dEngine

	return dEngine
}

func (de *decoratorEngine) Use(middleware ...gin.HandlerFunc) iDecoratorRoutes {
	filters, filtersHandlers := filterWrap(middleware)
	de.filters = append(de.filters, filters...)
	de.engine.Use(filtersHandlers...)

	return de
}
