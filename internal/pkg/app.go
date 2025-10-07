package pkg

import (
	"fmt"

	"lab3/internal/app/config"
	"lab3/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server starting up...")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)

	logrus.Infof("Server listening on %s", serverAddress)

	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}

	logrus.Info("Server shut down")
}
