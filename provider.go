package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Providers struct {
	Map map[string]struct{} `json:"map"`
}

func (p *Providers) Exists(name string) bool {
	_, ok := p.Map[name]
	return ok
}

func (p *Providers) Add(name string) error {
	if _, ok := p.Map[name]; ok {
		return fmt.Errorf("error. %s already exists as a provider", name)
	}

	p.Map[name] = struct{}{}
	return nil
}

func (p *Providers) Remove(name string) error {
	if _, ok := p.Map[name]; !ok {
		return fmt.Errorf("error. %s does not exist as a provider", name)
	}

	delete(p.Map, name)
	return nil
}

func (p *Providers) List() []string {
	providerList := []string{}

	for key := range p.Map {
		providerList = append(providerList, key)
	}

	return providerList
}

func (c *config) saveProviders() error {
	data, err := json.Marshal(c.Providers)
	if err != nil {
		return err
	}

	saveFile, err := os.OpenFile(c.pathToProviders, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() error {
		err = saveFile.Close()
		if err != nil {
			return err
		}
		return nil
	}()

	_, err = saveFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *config) PullProviders() error {
	_, err := os.Stat(c.pathToProviders)
	if err == nil {
		data, err := os.ReadFile(c.pathToProviders)
		if err != nil {
			return err
		}

		providers := Providers{}
		err = json.Unmarshal(data, &providers)
		if err != nil {
			c.Providers = providers
			return err
		}

		c.Providers = providers

	} else {
		return fmt.Errorf("warning: provider list not found")
	}

	return nil
}
