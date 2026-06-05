package pairmanager

import (
	"time"

	"FileSync/internal/models"
)

func Get(id int) (*models.SyncPair, error) {

	pairs, err := Load()
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {

		if pair.ID == id {
			return &pair, nil
		}
	}

	return nil, nil
}


func Delete(id int) error {

	pairs, err := Load()
	if err != nil {
		return err
	}

	var updated []models.SyncPair

	for _, pair := range pairs {

		if pair.ID != id {
			updated = append(updated, pair)
		}
	}

	return Save(updated)
}

func List() ([]models.SyncPair, error) {
	return Load()
}


func Add(source string, destination string) (models.SyncPair, error) {

	pairs, err := Load()
	if err != nil {
		return models.SyncPair{}, err
	}

	id := 1

	if len(pairs) > 0 {
		id = pairs[len(pairs)-1].ID + 1
	}

	pair := models.SyncPair{
		ID:          id,
		Source:      source,
		Destination: destination,
		CreatedAt:   time.Now(),
	}

	pairs = append(pairs, pair)

	err = Save(pairs)

	return pair, err
}


