package query

import (
	"errors"
	"strconv"
	"strings"
)

func validateGenre(genre string) error {

	if genre == "" {
		return errors.New("genre is not provided")
	}

	if _, exists := genresList[strings.ToLower(genre)]; !exists {
		return errors.New("genre is not correct")
	}

	return nil
}

func validateYear(year string) error {

	if year == "" {
		return errors.New("year is not provided")
	}

	_, err := strconv.Atoi(year)
	if err != nil {
		return errors.New("year is not integer")
	}
	return nil
}
