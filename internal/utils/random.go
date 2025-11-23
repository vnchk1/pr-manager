package utils

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"math/rand"
	"time"
)

func ShuffleUsers(users []*models.User) []*models.User {
	shuffled := make([]*models.User, len(users))
	copy(shuffled, users)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

// ShuffleStrings перемешивает слайс строк
func ShuffleStrings(items []string) []string {
	shuffled := make([]string, len(items))
	copy(shuffled, items)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}
