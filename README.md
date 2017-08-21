# Wait for Gitlab builds to finish

This is a command line tool for waiting for your Gitlab pipelines to finish.
Run `gitlab wait` and the CLI will find the most recent pipeline for your Git
branch, then wait for it to complete. If the build fails, `gitlab` will open a
browser tab to the job that failed.

Run `gitlab open` to open the most recent pipeline in your browser.

## Installation

Find your target operating system (darwin, windows, linux) and desired bin
directory, and modify the command below as appropriate:

    curl --silent --location https://github.com/kevinburke/gitlab/releases/download/0.3/gitlab-linux-amd64 > /usr/local/bin/gitlab && chmod 755 /usr/local/bin/gitlab

On Travis, you may want to create `$HOME/bin` and write to that, since
/usr/local/bin isn't writable with their container-based infrastructure.

The latest version is 0.3.

If you have a Go development environment, you can also install via source code:

    go get -u github.com/kevinburke/gitlab
