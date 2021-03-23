package froth

import (
	"bytes"
	"encoding/gob"
	"github.com/boggydigital/kvas"
)

type Stash struct {
	dst       string
	asset     string
	keyValues map[string]string
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

	var keyValues map[string]string

	if stashRC != nil {
		defer stashRC.Close()
		if err := gob.NewDecoder(stashRC).Decode(&keyValues); err != nil {
			return nil, err
		}
	}

	if keyValues == nil {
		keyValues = make(map[string]string, 0)
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

func (stash *Stash) Contains(id string) bool {
	_, ok := stash.keyValues[id]
	return ok
}

func (stash *Stash) Set(key string, value string) error {
	stash.keyValues[key] = value
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

func (stash *Stash) SetMany(keyValues map[string]string) error {
	for k, v := range keyValues {
		stash.keyValues[k] = v
	}
	return stash.write()
}

func (stash *Stash) Get(key string) (string, bool) {
	if stash == nil || stash.keyValues == nil {
		return "", false
	}
	val, ok := stash.keyValues[key]
	return val, ok
}
