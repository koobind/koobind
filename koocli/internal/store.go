package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"
)

type Config struct {
	Server string		`json:"server"`
	RootCaFile string 	`json:"rootCaFile"`
}

type CurrentContext struct {
	Context string	`json:"context"`
}

type TokenBag struct {
	Token      string    `json:"token"`
	ClientTTL  metav1.Duration  `json:"clientTTL"`
	LastAccess time.Time `json:"lastAccess"`
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

func getCurrentContextPath() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(usr.HomeDir, ".kube/cache/koo/context.json")
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


func LoadCurrentContext() string {
	currentContextPath := getCurrentContextPath()
	var currentContext CurrentContext
	if loadStuff(currentContextPath, func (decoder *json.Decoder) error {
		return decoder.Decode(&currentContext)
	}) {
		getLog().Debugf("LoadCurrentContext(%s) -> context:%s", currentContextPath, currentContext.Context)
		return currentContext.Context
	} else {
		getLog().Debugf("LoadCurrentContext(%s) -> ''", currentContextPath)
		return ""
	}
}

func LoadConfig(context string) *Config {
	configPath := getConfigPath(context)
	var config Config
	if loadStuff(configPath, func (decoder *json.Decoder) error {
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
	if loadStuff(tokenBagPath, func (decoder *json.Decoder) error {
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


func SaveCurrentContext(context string) {
	currentContextPath := getCurrentContextPath()
	getLog().Debugf("SaveCurrentContext(%s, %s)", currentContextPath, context)
	currentContext := CurrentContext{
		Context: context,
	}
	saveStuff(currentContextPath, func (encoder *json.Encoder) error  {
		return encoder.Encode(currentContext)
	})
}

func SaveConfig(context string, config *Config) {
	configPath := getConfigPath(context)
	getLog().Debugf("SaveConfig(%s, server:%s   rootCaFile:%s)", configPath, config.Server, config.RootCaFile)
	saveStuff(configPath, func (encoder *json.Encoder) error  {
		return encoder.Encode(config)
	})
}

func SaveTokenBag(context string, tokenBag *TokenBag) {
	tokenBagPath := getTokenBagPath(context)
	getLog().Debugf("SaveTokenBag(%s token:%s  ttl:%s  created:%s)", tokenBagPath, tokenBag.Token, tokenBag.ClientTTL, tokenBag.LastAccess)
	saveStuff(tokenBagPath, func (encoder *json.Encoder) error  {
		return encoder.Encode(tokenBag)
	})
}

// ----------------------------------------------------------------------------------

func loadStuff(path string, decode func (decoder *json.Decoder) error ) bool {
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



func saveStuff(path string, encode func (encoder *json.Encoder) error ) {
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


