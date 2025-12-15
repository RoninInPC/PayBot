package telegram

import (
	"fmt"
	"github.com/pkg/errors"
	"main/internal/database/cache"
	"strconv"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	AddRequisite    State = "add_requisite"
	AddPromocode    State = "add_promocode"
	AddResource     State = "add_resource"
	AddTariff       State = "add_tariff"
	ChangeRequisite State = "change_requisite"
	ChangePromocode State = "change_promocode"
	ChangeResource  State = "change_resource"
	ChangeTariff    State = "change_tariff"
)

type StateHem struct {
	sets cache.Sets
}

func InitState(sets cache.Sets) StateHem {
	return StateHem{sets: sets}
}

func (s StateHem) AddState(userID int64, state State, key, value string) error {
	userIdStr := strconv.FormatInt(userID, 10)
	contains, err := s.sets.Contains(userIdStr, state.String())
	if err != nil {
		return errors.Wrap(err, "Error checking state")
	}
	if !contains {
		err = s.sets.Clear(userIdStr)
		if err != nil {
			return errors.Wrap(err, "Error clearing state")
		}
		err = s.sets.Add(userIdStr, state.String())
		if err != nil {
			return errors.Wrap(err, "Error adding state")
		}
	}
	err = s.sets.Add(userIdStr, fmt.Sprintf(key, ":", value))
	if err != nil {
		return errors.Wrap(err, "Error adding key and value")
	}
	return nil
}

func (s StateHem) GetValues(userID int64) ([]string, error) {
	userIdStr := strconv.FormatInt(userID, 10)
	ans, err := s.sets.GetAll(userIdStr)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting values")
	}
	return ans, nil
}

func (s StateHem) Clear(userID int64) error {
	userIdStr := strconv.FormatInt(userID, 10)
	err := s.sets.Clear(userIdStr)
	if err != nil {
		return errors.Wrap(err, "Error clearing state")

	}
	return nil
}
