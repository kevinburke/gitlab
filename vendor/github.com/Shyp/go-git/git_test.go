package git

import (
	"fmt"
	"testing"
)

func TestCurrentBranch(t *testing.T) {
	result, err := CurrentBranch()
	fmt.Println(err)
	fmt.Println(result)
	fmt.Println(len(result))
}

var remoteTests = []struct {
	remote   string
	expected RemoteURL
}{
	{
		"git@github.com:Shyp/shyp_api.git", RemoteURL{
			Host:     "github.com",
			Port:     22,
			Path:     "Shyp",
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:Shyp/shyp_api.git",
			SSHUser:  "git",
		},
	}, {
		"git@github.com:Shyp/shyp_api.git/", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     22,
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:Shyp/shyp_api.git/",
			SSHUser:  "git",
		},
	}, {
		"git@github.com:path/to/Shyp/shyp_api.git/", RemoteURL{
			Path:     "path/to/Shyp",
			Host:     "github.com",
			Port:     22,
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:path/to/Shyp/shyp_api.git/",
			SSHUser:  "git",
		},
	}, {
		"https://github.com/Shyp/shyp_api.git", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/shyp_api.git",
			SSHUser:  "",
		},
	}, {
		"https://github.com/Shyp/shyp_api.git/", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/shyp_api.git/",
			SSHUser:  "",
		},
	}, {
		"https://github.com:11443/Shyp/shyp_api.git", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     11443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com:11443/Shyp/shyp_api.git",
			SSHUser:  "",
		},
	},
}

func TestParseRemoteURL(t *testing.T) {
	for _, tt := range remoteTests {
		remote, err := ParseRemoteURL(tt.remote)
		if err != nil {
			t.Fatal(err)
		}
		if remote == nil {
			t.Fatalf("expected ParseRemoteURL(%s) to be %v, was nil", tt.remote, tt.expected)
		}
		if *remote != tt.expected {
			t.Errorf("expected ParseRemoteURL(%s) to be %#v, was %#v", tt.remote, tt.expected, remote)
		}
	}
}
