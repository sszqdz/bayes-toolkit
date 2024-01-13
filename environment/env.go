// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package environment

import (
	"fmt"
	"os"
	"sync"

	"github.com/sszqdz/bayes-toolkit/dirr"

	"github.com/spf13/viper"
)

type ENV string

const (
	EnvDebug   ENV = "debug"
	EnvRelease ENV = "release"
)

var (
	ENVKey   = "ENV"
	FileName = ".env"
	Deepth   = 3
	loadEnv  = sync.OnceValue[string](extractEnv)
	loadFile = sync.OnceValue[*viper.Viper](loadFileConf)
)

// do not auto inject, use when runtime
func LoadEnv() ENV {
	return ENV(loadEnv())
}

func (e ENV) Is(env ENV) bool {
	return e == env
}

func (e ENV) String() string {
	return string(e)
}

func Load(key string) string {
	return extractKey(key)
}

func extractEnv() string {
	return extractKey(ENVKey)
}

func extractKey(key string) string {
	// Extract environment variable
	env := os.Getenv(key)
	if len(env) > 0 {
		return env
	}
	// Extract file variable
	vFile := loadFile()
	if vFile != nil {
		return vFile.GetString(key)
	}

	return ""
}

func loadFileConf() *viper.Viper {
	// Upward recursive search for .env files
	curDir, err := os.Getwd()
	panicErr(err)
	fmt.Println("curDir: " + curDir)
	envFilePath, err := dirr.FindFileInParentDirs(FileName, curDir, Deepth)
	fmt.Println("envFilePath: " + envFilePath)
	if os.IsNotExist(err) {
		return nil
	}
	panicErr(err)

	vFile := viper.New()
	vFile.SetConfigFile(envFilePath)
	panicErr(vFile.ReadInConfig())
	return vFile
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
