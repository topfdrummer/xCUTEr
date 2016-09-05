// Copyright (c) 2016 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.

package job

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

type templatingEngine struct {
	Config *Config
	Host   *host
	Env    map[string]string
	now    func() time.Time
}

func getEnv() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		index := strings.Index(e, "=")
		key, value := e[:index], e[index+1:]
		env[key] = value
	}

	fmt.Printf("%#v\n", env)

	return env
}

func newTemplatingEngine(c *Config, h *host) *templatingEngine {
	return &templatingEngine{
		Config: c,
		Host:   h,
		Env:    getEnv(),
		now:    time.Now,
	}
}

func (t *templatingEngine) Interpolate(templ string) (string, error) {
	var buf bytes.Buffer

	funcMap := template.FuncMap{
		"date": func(t time.Time) string {
			return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
		},
		"time": func(t time.Time) string {
			return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
		},
		"now": func() time.Time {
			return t.now()
		},
	}

	tt := template.New("").Funcs(funcMap)

	tt, err := tt.Parse(templ)
	if err != nil {
		return "", err
	}

	data := struct {
		Config *Config
		Host   *host
		Env    map[string]string
		Now    time.Time
	}{
		Config: t.Config,
		Host:   t.Host,
		Env:    t.Env,
		Now:    time.Now(),
	}

	err = tt.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
