package monitor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/sirupsen/logrus"
)

type healthHandler struct {
	esClient *es.ESClient
}

func NewHealthHandler(esClient *es.ESClient) (*healthHandler, error) {

	return &healthHandler{
		esClient: esClient,
	}, nil

}

func (m *healthHandler) Routes(app *fiber.App, router fiber.Router) {

	router.Get("/report", m.getHealthStatusReport)
	router.Get("/status", m.getHealthStatus)

}

func (m *healthHandler) getHealthStatus(c *fiber.Ctx) error {

	logrus.Infof("getHealthStatus ")

	err := m.esClient.Ping()
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	return c.SendStatus(fiber.StatusOK) // Send 200 OK with no content

}

func (m *healthHandler) getHealthStatusReport(c *fiber.Ctx) error {

	logrus.Infof("getHealthStatusReport ")

	esState := &ServiceState{
		Healthy: true,
	}
	esErr := m.esClient.Ping()
	if esErr != nil {
		esState.Error = esErr.Error()
		esState.Healthy = false
	}

	return c.JSON(&StatusReport{
		ElasticSearch: esState,
	})

}
