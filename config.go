package weddingseats

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	Facebook struct {
		ClientId     string
		ClientSecret string
		AuthURL      string
		TokenURL     string
		RedirectURL  string
	}
}

func ReadConfig(path string) (err error) {
	local_conf := new(Configuration)
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	log.Printf("loaded config from %s", path)
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(local_conf)
	if err != nil {
		return err
	}
	configLock.Lock()
	config = local_conf
	configLock.Unlock()
	return nil
}
