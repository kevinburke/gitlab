package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Shyp/go-git"
	types "github.com/Shyp/go-types"
	"github.com/kevinburke/rest"
)

const Version = "0.5"

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", msg, err)
		os.Exit(1)
	}
}

type ListPipeline struct {
	ID     uint64 `json:"id"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
	Status string `json:"status"`
}

type Job struct {
	ID         uint64         `json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	FinishedAt types.NullTime `json:"finished_at"`
	Name       string         `json:"name"`
	StartedAt  types.NullTime `json:"started_at"`
	Stage      string         `json:"stage"`
	Status     string         `json:"status"`
}

func getJobs(ctx context.Context, client *rest.Client, cfg *PipelineConfig) ([]*Job, error) {
	path := fmt.Sprintf("/projects/%s%%2F%s/pipelines/%d/jobs?private_token=%s", cfg.Org, cfg.RepoName, cfg.ID, cfg.Token)
	req, err := client.NewRequest("GET", path, nil)
	checkError(err, "creating request")
	jobs := make([]*Job, 0)
	req = req.WithContext(ctx)
	doErr := client.Do(req, &jobs)
	return jobs, doErr
}

// {"id":10903,"sha":"c6c403a54187346da1807364b6a42267af1cf686","ref":"fix-tests","status":"running","before_sha":"c6c403a54187346da1807364b6a42267af1cf686","tag":false,"yaml_errors":null,"user":{"name":"Kevin
// Burke","username":"kb","id":54,"state":"active","avatar_url":"https://secure.gravatar.com/avatar/7e614caf6d6d0379ef1f6d05f4cdb39d?s=80&d=identicon","web_url":"anurl"},"created_at":"2017-08-19T01:43:52.486Z","updated_at":"2017-08-19T01:43:54.143Z","started_at":"2017-08-19T01:43:54.141Z","finished_at":null,"committed_at":null,"duration":null,"coverage":null}
type Pipeline struct {
	ID        uint64    `json:"id"`
	Ref       string    `json:"ref"`
	SHA       string    `json:"sha"`
	Status    string    `json:"status"`
	BeforeSHA string    `json:"before_sha"`
	Tag       bool      `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	StartedAt time.Time `json:"started_at"`
}

func getLatestPipeline(ctx context.Context, cfg *PipelineConfig) (*ListPipeline, error) {
	branch, err := git.CurrentBranch()
	if err != nil {
		return nil, err
	}
	cfg.Branch = branch
	client := rest.NewClient("", "", "https://"+cfg.Host+"/api/v4")
	path := fmt.Sprintf("/projects/%s%%2F%s/pipelines?private_token=%s", cfg.Org, cfg.RepoName, cfg.Token)
	req, err := client.NewRequest("GET", path, nil)
	checkError(err, "creating request")
	results := make([]*ListPipeline, 0)
	req = req.WithContext(ctx)
	doErr := client.Do(req, &results)
	checkError(doErr, "getting pipelines")
	var pipeline *ListPipeline
	for i := range results {
		if results[i].Ref == branch {
			pipeline = results[i]
			break
		}
	}
	if pipeline == nil {
		checkError(errors.New("could not find pipeline for that branch"), "getting pipelines")
	}
	return pipeline, nil
}

type PipelineConfig struct {
	Host     string
	RepoName string
	Org      string
	ID       uint64
	Token    string
	Branch   string
}

func openPipeline(ctx context.Context, cfg *PipelineConfig) error {
	path := fmt.Sprintf("https://%s/%s/%s/pipelines/%d", cfg.Host, cfg.Org, cfg.RepoName, cfg.ID)
	cmd := exec.CommandContext(ctx, "open", path)
	return cmd.Run()
}

func openJob(ctx context.Context, cfg *PipelineConfig, job *Job) error {
	path := fmt.Sprintf("https://%s/%s/%s/-/jobs/%d", cfg.Host, cfg.Org, cfg.RepoName, job.ID)
	cmd := exec.CommandContext(ctx, "open", path)
	return cmd.Run()
}

func open(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	token, err := getToken(remote.Host)
	checkError(err, "getting token")
	cfg := &PipelineConfig{
		Host:     remote.Host,
		RepoName: remote.RepoName,
		Org:      remote.Path,
		Token:    token,
	}
	pipeline, err := getLatestPipeline(ctx, cfg)
	if err != nil {
		return err
	}
	cfg.ID = pipeline.ID
	return openPipeline(ctx, cfg)
}

func wait(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	token, err := getToken(remote.Host)
	checkError(err, "getting token")
	cfg := &PipelineConfig{
		Host:     remote.Host,
		RepoName: remote.RepoName,
		Org:      remote.Path,
		Token:    token,
	}
	client := rest.NewClient("", "", "https://"+remote.Host+"/api/v4")
	pipeline, err := getLatestPipeline(ctx, cfg)
	if err != nil {
		return err
	}
	cfg.ID = pipeline.ID
	for {
		path := fmt.Sprintf("/projects/%s%%2F%s/pipelines/%d?private_token=%s", remote.Path, remote.RepoName, pipeline.ID, token)
		req, err := client.NewRequest("GET", path, nil)
		checkError(err, "creating request")
		req = req.WithContext(ctx)
		pipeline := new(Pipeline)
		doErr := client.Do(req, pipeline)
		checkError(doErr, "getting pipeline")
		jobs, err := getJobs(ctx, client, cfg)
		checkError(err, "getting pipeline jobs")
		if pipeline.Status == "failed" {
			os.Stdout.WriteString("pipeline failed\n")
			for _, job := range jobs {
				if job.Status == "failed" {
					openJob(ctx, cfg, job)
					fmt.Fprintf(os.Stderr, "https://%s/%s/%s/-/jobs/%d\n", cfg.Host, cfg.Org, cfg.RepoName, job.ID)
					os.Exit(1)
				}
			}
			openPipeline(ctx, cfg)
			fmt.Fprintf(os.Stderr, "https://%s/%s/%s/pipelines/%d\n", cfg.Host, cfg.Org, cfg.RepoName, cfg.ID)
			os.Exit(1)
		}
		if pipeline.Status == "success" {
			os.Stdout.WriteString("pipeline succeeded!\n")
			os.Exit(0)
		}
		d := time.Since(pipeline.CreatedAt)
		if d < time.Minute {
			d = d.Round(500 * time.Millisecond)
		} else {
			d = d.Round(time.Second)
		}
		foundJob := false
		for i, job := range jobs {
			if job.Status == "running" {
				var d2 time.Duration
				if job.StartedAt.Valid && !job.StartedAt.Time.IsZero() {
					d2 = time.Since(job.StartedAt.Time)
				} else {
					d2 = time.Since(job.CreatedAt)
				}
				if d2 < time.Minute {
					d2 = d2.Round(500 * time.Millisecond)
				} else {
					d2 = d2.Round(time.Second)
				}
				fmt.Printf("Job #%d (%q step %d) is %s for %s, %s since push, sleeping...\n", job.ID, job.Name, i+1, job.Status, d2, d)
				foundJob = true
				break
			}
		}
		time.Sleep(3 * time.Second)
		if foundJob {
			continue
		}
		fmt.Printf("Status is %s, %s since push, sleeping...\n", pipeline.Status, d)
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		os.Stderr.WriteString("please provide an argument\n")
		os.Exit(2)
	}
	args := flag.Args()
	switch args[0] {
	case "help":
		os.Stderr.WriteString("gitlab [wait|help|open|version]\n")
		os.Exit(2)
	case "version":
		os.Stderr.WriteString("gitlab version " + Version + "\n")
		os.Exit(2)
	case "wait":
		wait(args[1:])
	case "open":
		open(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown argument %s\n", args[0])
		os.Exit(2)
	}
}
