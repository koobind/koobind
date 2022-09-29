/*
Copyright (C) 2020 Serge ALEXANDRE

# This file is part of koobind project

koobind is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

koobind is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
package internal

import (
	"encoding/json"
	"fmt"
	proto_v2 "github.com/koobind/koobind/koomgr/apis/proto/auth/v2"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"
)

type Config struct {
	Server     string `json:"server"`
	RootCaFile string `json:"rootCaFile"`
	Client     proto_v2.AuthClient
}

type CurrentContext struct {
	Context string `json:"context"`
}

type TokenBag struct {
	Token      string        `json:"token"`
	ClientTTL  time.Duration `json:"clientTTL"`
	LastAccess time.Time     `json:"lastAccess"`
	Username   string        `json:"username"`
	Uid        string        `json:"uid"`
	Groups     []string      `json:"groups"`
}

func getConfigPath(context string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, fmt.Sprintf(".kube/cache/koo/%s/config.json", context))
}

func getTokenBagPath(context string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, fmt.Sprintf(".kube/cache/koo/%s/tokenbag.json", context))
}

func ListContext() []string {
	files := make([]string, 0, 10)
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	fl, err := ioutil.ReadDir(path.Join(usr.HomeDir, ".kube/cache/koo/"))
	for _, f := range fl {
		if f.IsDir() {
			files = append(files, f.Name())
		}
	}
	return files
}

func LoadConfig(context string) *Config {
	configPath := getConfigPath(context)
	var config Config
	if loadStuff(configPath, func(decoder *json.Decoder) error {
		return decoder.Decode(&config)
	}) {
		getLog().Debugf("LoadConfig(%s) -> server:%s  rootCaFile:%s", configPath, config.Server, config.RootCaFile)
		return &config
	} else {
		getLog().Debugf("LoadConfig(%s) -> nil", configPath)
		return nil
	}
}

func LoadTokenBag(context string) *TokenBag {
	tokenBagPath := getTokenBagPath(context)
	var tokenBag TokenBag
	if loadStuff(tokenBagPath, func(decoder *json.Decoder) error {
		return decoder.Decode(&tokenBag)
	}) {
		getLog().Debugf("LoadTokenBag(%s) -> token:%s  ttl:%s  created:%s", tokenBagPath, tokenBag.Token, tokenBag.ClientTTL, tokenBag.LastAccess)
		return &tokenBag
	} else {
		getLog().Debugf("LoadTokenBag(%s) -> nil", tokenBagPath)
		return nil
	}
}

// Better to test and remove. Alternate would be to remove withhout testing, but this may hide some errors
func DeleteTokenBag(context string) {
	tokenBagPath := getTokenBagPath(context)
	getLog().Debugf("DeleteTokenBag(%s)", tokenBagPath)
	_, err := os.Stat(tokenBagPath)
	if !os.IsNotExist(err) {
		err := os.Remove(tokenBagPath)
		if err != nil {
			panic(err)
		}
	}
	return
}

func SaveConfig(context string, config *Config) {
	configPath := getConfigPath(context)
	getLog().Debugf("SaveConfig(%s, server:%s   rootCaFile:%s)", configPath, config.Server, config.RootCaFile)
	saveStuff(configPath, func(encoder *json.Encoder) error {
		return encoder.Encode(config)
	})
}

func SaveTokenBag(context string, tokenBag *TokenBag) {
	tokenBagPath := getTokenBagPath(context)
	getLog().Debugf("SaveTokenBag(%s token:%s  ttl:%s  created:%s)", tokenBagPath, tokenBag.Token, tokenBag.ClientTTL, tokenBag.LastAccess)
	saveStuff(tokenBagPath, func(encoder *json.Encoder) error {
		return encoder.Encode(tokenBag)
	})
}

// ----------------------------------------------------------------------------------

func loadStuff(path string, decode func(decoder *json.Decoder) error) bool {
	if file, err := os.Open(path); err == nil {
		err = decode(json.NewDecoder(file))
		if err != nil {
			panic(err)
		}
		_ = file.Close()
		return true
	} else {
		return false
	}
}

func saveStuff(path string, encode func(encoder *json.Encoder) error) {
	ensureDir(filepath.Dir(path))
	var err error
	var file *os.File
	if file, err = os.Create(path); err == nil {
		if err = encode(json.NewEncoder(file)); err == nil {
			err = file.Close()
		}
	}
	if err != nil {
		panic(err)
	}

}

func ensureDir(dirName string) {
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, 0700)
		if merr != nil {
			panic(merr)
		}
	}
}
