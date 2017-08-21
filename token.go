package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

type GitlabConfig struct {
	Organizations map[string]Organization
}

type Organization struct {
	Host  string
	Token string
}

func getToken(orgName string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	filename := filepath.Join(user.HomeDir, "cfg", "gitlab")
	f, err := os.Open(filename)
	rcFilename := ""
	if err != nil {
		rcFilename = filepath.Join(user.HomeDir, ".gitlabrc")
		f, err = os.Open(rcFilename)
	}
	if err != nil {
		err = fmt.Errorf(`Couldn't find a config file in %s or %s.

Add a configuration file with your Gitlab token, like this:

[organizations]

    [organizations.mycompany]
    host = "gitlab.mycompany.com"
    token = "aabbccddeeff00"

Go to https://gitlab.<mycompany>.com/profile/personal_access_tokens if you need 
to create a token.
`, filename, rcFilename)

		return "", err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	c := new(GitlabConfig)
	err = toml.Unmarshal(data, c)
	if err != nil {
		return "", err
	}
	for _, org := range c.Organizations {
		if org.Host == orgName {
			return org.Token, nil
		}
	}
	return "", fmt.Errorf(`Could not find host %[1]s in the config.

Add a token at %[1]s/profile/personal_access_tokens, then add it to the 
config file.
`, orgName)
}
