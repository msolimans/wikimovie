package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/sirupsen/logrus"
)

const moviesIndexName = "my-index"

// const moviesIndexName = "movies"

type movieHandler struct {
	esClient *es.ESClient
}

func NewMovieHandler(esClient *es.ESClient) (*movieHandler, error) {

	return &movieHandler{
		esClient: esClient,
	}, nil

}

func (m *movieHandler) Routes(app *fiber.App, movie fiber.Router) {

	movie.Get("/year/:year", m.getByYear)
	movie.Get("/genre/:genre", m.getByGenre)
	///if any other funcs related to cast then I would go with cast/{x}/movies/ for separation of concern
	movie.Get("/cast/:cast", m.getByCast)
	movie.Get("/query", m.searchMovie)

	movie.Get("/:title", m.searchByTitle)
}

func (m *movieHandler) getByYear(c *fiber.Ctx) error {
	year := c.Params("year")
	logrus.Infof("getByYear %v", year)
	fyear := strings.Trim(year, " ")

	if err := validateYear(fyear); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	query := fmt.Sprintf(`{
		"query": {
		  "match": {
			"year": "%v"
		  }
		}
	  }`, fyear)

	docs, err := m.esClient.QueryDocuments(moviesIndexName, query)
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	// res, err := json.Marshal(docs)

	// if err != nil {
	// 	c.SendStatus(500)
	// 	return c.SendString("Error in preparing movies list, please try again!")
	// }

	c.Set("Content-Type", "application/json")

	if len(docs) == 0 {
		return c.JSON([]interface{}{})
	}

	return c.JSON(docs)
}

func (m *movieHandler) getByGenre(c *fiber.Ctx) error {
	genre := c.Params("genre")

	logrus.Infof("getByGenre %v", genre)

	fgenre := strings.Trim(genre, " ")

	if err := validateGenre(fgenre); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	///////

	//wildcard instead of terms for partial matches
	//terms will match any f the values in the array
	query := fmt.Sprintf(`{
		"query": {
		  "terms": {
			"genres": ["%v"]
		  }
		}
	  }`, fgenre)

	docs, err := m.esClient.QueryDocuments(moviesIndexName, query)
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	c.Set("Content-Type", "application/json")
	if len(docs) == 0 {
		return c.JSON([]interface{}{})
	}
	return c.JSON(docs)

}

func (m *movieHandler) getByCast(c *fiber.Ctx) error {
	cast := c.Params("cast")
	logrus.Infof("getByCast %v", cast)
	query := fmt.Sprintf(`{
		"query": {
		  "terms": {
			"genres": ["%v"]
		  }
		}
	  }`, cast)

	docs, err := m.esClient.QueryDocuments(moviesIndexName, query)
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	c.Set("Content-Type", "application/json")
	if len(docs) == 0 {
		return c.JSON([]interface{}{})
	}
	return c.JSON(docs)

}

func (m *movieHandler) searchByTitle(c *fiber.Ctx) error {
	title := c.Params("title")
	logrus.Infof("searchByTitle %v", title)
	query := fmt.Sprintf(`{
		"query": {
		  "match": {
			"title": "%v"
		  }
		}
	  }`, title)

	docs, err := m.esClient.QueryDocuments(moviesIndexName, query)
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	c.Set("Content-Type", "application/json")
	if len(docs) == 0 {
		return c.JSON([]interface{}{})
	}
	return c.JSON(docs)

}

func (m *movieHandler) searchMovie(c *fiber.Ctx) error {
	// Handler logic to delete a movie
	// Get the query parameters from the request context
	title := c.Query("title")
	genre := c.Query("genre")
	cast := c.Query("cast")
	year := c.Query("year")

	logrus.Infof("searchMovie %v %v %v %v", title, genre, cast, year)

	fyear := strings.Trim(year, " ")
	fgenre := strings.Trim(genre, " ")

	computed := []map[string]interface{}{}

	if title != "" {
		computed = append(computed, map[string]interface{}{
			"match": map[string]interface{}{
				"title": title,
			},
		})
	}

	if fyear != "" {
		if err := validateYear(year); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		computed = append(computed, map[string]interface{}{
			"match": map[string]interface{}{
				"year": fyear,
			},
		})
	}

	if fgenre != "" {
		if err := validateGenre(genre); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		computed = append(computed, map[string]interface{}{
			"terms": map[string]interface{}{
				"genres": []string{fgenre},
			},
		})
	}

	if cast != "" {
		computed = append(computed, map[string]interface{}{
			"terms": map[string]interface{}{
				"casts": []string{cast},
			},
		})
	}

	computedStr, err := json.Marshal(computed)

	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	query := fmt.Sprintf(`{
		"query": {
		  "bool": {
			"must": %s
		  }
		}
	  }
	  `, computedStr)

	docs, err := m.esClient.QueryDocuments(moviesIndexName, query)
	if err != nil {
		c.SendStatus(500)
		return c.SendString("Error retrieving movies, please try again!")
	}

	c.Set("Content-Type", "application/json")
	if len(docs) == 0 {
		return c.JSON([]interface{}{})
	}
	return c.JSON(docs)
}
