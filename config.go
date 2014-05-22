package weddingseats

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	CookieSecret string
	Facebook     struct {
		ClientId     string
		ClientSecret string
		AuthURL      string
		TokenURL     string
		RedirectURL  string
		Scope        string
	}
}

func ReadConfig(path string) (err error) {
	local_conf := new(Configuration)
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
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
