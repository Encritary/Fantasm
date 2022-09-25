package main

import (
	"fmt"
	"github.com/Encritary/Fantasm/fantasm"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log := logrus.New()
	log.Level = logrus.DebugLevel

	log.Info("Loading config...")

	configRaw, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}

	config := &Config{}
	err = yaml.Unmarshal(configRaw, config)
	if err != nil {
		panic(err)
	}

	for _, template := range config.ActionTemplates {
		if template.Template != "" {
			panic("templates in templates are not supported")
		}
	}

	log.Info("Loading sections...")

	sections := make(map[string]*fantasm.Section)

	err = filepath.Walk("sections", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".yml") || strings.HasSuffix(info.Name(), ".yaml") {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			var sectionsMap map[string]*fantasm.Section
			err = yaml.Unmarshal(data, &sectionsMap)
			if err != nil {
				fmt.Println(path)
				panic(err)
			}

			for id, sec := range sectionsMap {
				if _, ok := sections[id]; ok {
					panic("section ID conflict")
				}

				sec.ID = id
				for _, action := range sec.Actions {
					if action.Template != "" {
						if parent, ok := config.ActionTemplates[action.Template]; ok {
							if action.Label == "" {
								action.Label = parent.Label
							}
							if action.Color == "" {
								action.Color = parent.Color
							}
							if action.Section == "" {
								action.Section = strings.ReplaceAll(parent.Section, "{current}", sec.ID)
							}
						} else {
							panic("unknown template " + action.Template)
						}
					}
				}
				sections[id] = sec
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, sec := range sections {
		for _, action := range sec.Actions {
			if action.Section != "__back" {
				if _, ok := sections[action.Section]; !ok {
					panic("section " + action.Section + " doesn't exist")
				}
			}
		}
	}

	log.Info("Starting bot...")

	fsm, err := fantasm.NewFastasm(log, config.VKToken, sections)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for events")
	fsm.Run()
}

type Config struct {
	VKToken         string                    `yaml:"vk_token"`
	ActionTemplates map[string]fantasm.Action `yaml:"action_templates"`
}
