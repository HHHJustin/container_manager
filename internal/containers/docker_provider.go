package containers

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type DockerProvider struct{ cli *client.Client }

func NewDockerProvider() *DockerProvider {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &DockerProvider{cli: cli}
}

func (d *DockerProvider) Create(opts CreateOptions) (Container, error) {
	ctx := context.Background()

	if rc, err := d.cli.ImagePull(ctx, opts.Image, image.PullOptions{}); err == nil {
		// Drain and close to avoid leaking the connection; ignore content
		_, _ = io.Copy(io.Discard, rc)
		_ = rc.Close()
	}

	name := opts.Name
	resp, err := d.cli.ContainerCreate(ctx, &container.Config{Image: opts.Image}, nil, nil, nil, name)
	if err != nil {
		return Container{}, err
	}

	c := Container{
		ID:        resp.ID,
		Name:      name,
		Image:     opts.Image,
		CreatedAt: time.Now().Unix(),
		Status:    "created",
	}
	return c, nil
}

func (d *DockerProvider) Start(id string) error {
	ctx := context.Background()
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *DockerProvider) Stop(id string) error {
	ctx := context.Background()
	timeout := 10 * time.Second
	seconds := int(timeout.Seconds())
	return d.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &seconds})
}

func (d *DockerProvider) Delete(id string) error {
	ctx := context.Background()
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

// RunJob 實作一次性作業：綁定 host 資料夾並執行命令，回傳退出碼與日誌。
func (d *DockerProvider) RunJob(opts JobOptions) (int64, string, error) {
	ctx := context.Background()
	if rc, err := d.cli.ImagePull(ctx, opts.Image, image.PullOptions{}); err == nil {
		_, _ = io.Copy(io.Discard, rc)
		_ = rc.Close()
	}

	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:      opts.Image,
		Cmd:        opts.Cmd,
		WorkingDir: opts.ContainerDir,
		Tty:        false,
	}, &container.HostConfig{
		Mounts: []mount.Mount{{Type: mount.TypeBind, Source: opts.HostDir, Target: opts.ContainerDir, ReadOnly: false}},
	}, nil, nil, "")
	if err != nil {
		return 0, "", err
	}
	id := resp.ID
	defer func() { _ = d.cli.ContainerRemove(context.Background(), id, container.RemoveOptions{Force: true}) }()

	if err := d.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return 0, "", err
	}
	// 等待結束
	statusCh, errCh := d.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case st := <-statusCh:
		// 讀取輸出
		logsRc, err := d.cli.ContainerLogs(ctx, id, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: false})
		if err != nil {
			return st.StatusCode, "", nil
		}
		defer logsRc.Close()
		b, _ := io.ReadAll(logsRc)
		return st.StatusCode, string(b), nil
	case err := <-errCh:
		return 0, "", err
	case <-time.After(10 * time.Minute):
		_ = d.cli.ContainerStop(ctx, id, container.StopOptions{})
		return -1, "timeout", nil
	}
}

var _ JobRunner = (*DockerProvider)(nil)
