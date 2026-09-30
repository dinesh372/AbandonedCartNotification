package repository

import (
	"cmp"
	"fmt"
	"slices"
)

type SortedSetEntry struct {
	score int64
	key   string
}

func (e SortedSetEntry) Score() int64 {
	return e.score
}

func (e SortedSetEntry) Key() string {
	return e.key
}

type SortedSet interface {
	Store(score int64, key string) error
	GetScore(key string) (*int64, error)
	Delete(key string) error
	ConditionalDelete(key string, value int64) error
	Dequeue(maxScore int64) (*SortedSetEntry, error)
}

type sortedSet struct {
	items []SortedSetEntry
	name  string
}

func NewSortedSet(name string) SortedSet {
	return &sortedSet{
		name: name,
	}
}

func sortSlice(items []SortedSetEntry) {
	slices.SortFunc(items, func(a, b SortedSetEntry) int {
		return cmp.Compare(a.score, b.score)
	})
}

func (s *sortedSet) Store(score int64, key string) error {
	alreadyPresent := false
	newEntry := SortedSetEntry{
		score: score,
		key:   key,
	}

	for i, entry := range s.items {
		if entry.key == key {
			fmt.Printf(s.name+" Updating an existing entry old = %+v  new = %+v \n", entry, newEntry)
			// replace older value in place.
			s.items[i] = newEntry
			alreadyPresent = true
			break
		}
	}

	// Add new entry as not present already.
	if !alreadyPresent {
		fmt.Printf(s.name+" Adding a new entry new = %+v \n", newEntry)
		s.items = append(s.items, newEntry)
	}

	sortSlice(s.items)
	return nil
}

func (s *sortedSet) GetScore(key string) (*int64, error) {
	for _, entry := range s.items {
		if entry.key == key {
			return &entry.score, nil
		}
	}
	return nil, nil
}

func (s *sortedSet) Delete(key string) error {

	length := len(s.items)

	for i, entry := range s.items {
		if entry.key == key {
			fmt.Printf(s.name+" Deleting an entry old = %+v \n", entry)

			// replace with last element & delete last element
			s.items[i] = s.items[length-1]
			s.items = s.items[:length-1]
			break
		}
	}
	sortSlice(s.items)
	return nil
}

func (s *sortedSet) ConditionalDelete(key string, score int64) error {
	for _, entry := range s.items {
		if entry.key == key && entry.score == score {
			fmt.Printf(s.name+" Conditionally deleting an entry old = %+v\n", entry)
			s.Delete(key)
		}
	}
	return nil
}

/** Dequeue - returns the lowest score entry in the set which is not greater than maxScore.
 * This also removes the entry from the sorted set.
 * For schedules, pass time.Now().Unix() so due tasks (scheduledAt <= now) are claimed.
 */
func (s *sortedSet) Dequeue(maxScore int64) (*SortedSetEntry, error) {
	if s.isEmpty() {
		return nil, nil
	}

	minScoreEntry := s.items[0]
	if minScoreEntry.score <= maxScore {
		fmt.Printf(s.name+" Dequeing an entry, entry = %+v  maxscore = %+v \n", minScoreEntry, maxScore)

		length := len(s.items)
		// removes the first entry from sorted set
		s.items = s.items[1:length]
		return &minScoreEntry, nil
	}

	return nil, nil

}

func (s *sortedSet) isEmpty() bool {
	return len(s.items) == 0
}
