package froth

import (
	"bytes"
	"encoding/gob"
	"github.com/boggydigital/kvas"
	"strings"
)

type Stash struct {
	dst       string
	asset     string
	keyValues map[string][]string
}

func NewStash(dst, asset string) (*Stash, error) {
	kvStash, err := kvas.NewGobLocal(dst)
	if err != nil {
		return nil, err
	}

	stashRC, err := kvStash.Get(asset)
	if err != nil {
		return nil, err
	}

	var keyValues map[string][]string

	if stashRC != nil {
		defer stashRC.Close()
		if err := gob.NewDecoder(stashRC).Decode(&keyValues); err != nil {
			return nil, err
		}
	}

	if keyValues == nil {
		keyValues = make(map[string][]string, 0)
	}

	return &Stash{
		dst:       dst,
		asset:     asset,
		keyValues: keyValues,
	}, nil
}

func (stash *Stash) All() []string {
	keys := make([]string, 0, len(stash.keyValues))
	for k := range stash.keyValues {
		keys = append(keys, k)
	}
	return keys
}

func (stash *Stash) Contains(key string) bool {
	_, ok := stash.keyValues[key]
	return ok
}

func (stash *Stash) ContainsValue(key string, value string) bool {
	for _, val := range stash.keyValues[key] {
		if val == value {
			return true
		}
	}
	return false
}

func (stash *Stash) Add(key string, value string) error {
	if stash.ContainsValue(key, value) {
		return nil
	}
	stash.keyValues[key] = append(stash.keyValues[key], value)
	return stash.write()
}

func (stash *Stash) write() error {
	kvStash, err := kvas.NewGobLocal(stash.dst)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(stash.keyValues); err != nil {
		return err
	}

	return kvStash.Set(stash.asset, buf)
}

func (stash *Stash) AddMany(keyValues map[string][]string) error {
	for key, values := range keyValues {
		for _, val := range values {
			if stash.ContainsValue(key, val) {
				continue
			}
			stash.keyValues[key] = append(stash.keyValues[key], val)
		}
	}
	return stash.write()
}

func (stash *Stash) Get(key string) (string, bool) {
	values, ok := stash.GetAll(key)
	if len(values) > 0 {
		return values[0], ok
	}
	return "", false
}

func (stash *Stash) GetAll(key string) ([]string, bool) {
	if stash == nil || stash.keyValues == nil {
		return nil, false
	}
	val, ok := stash.keyValues[key]
	return val, ok
}

func (stash *Stash) Search(term string, anyCase bool) []string {
	if anyCase {
		term = strings.ToLower(term)
	}
	matchingKeys := make([]string, 0)
	for key, values := range stash.keyValues {
		for _, val := range values {
			if anyCase {
				val = strings.ToLower(val)
			}
			if strings.Contains(val, term) {
				matchingKeys = append(matchingKeys, key)
			}
		}
	}
	return matchingKeys
}
