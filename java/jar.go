package wallepack

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Packer using the CMD command to call jar, so need java environment.
type Packer struct {
	jarPath  string
	_extract *regexp.Regexp
}

func WithJarPath(path string) func(*Packer) {
	return func(p *Packer) {
		p.jarPath = path
	}
}

func NewPacker(opts ...func(*Packer)) (*Packer, error) {
	packer := &Packer{jarPath: "", _extract: regexp.MustCompile(`([\w.]+)=([\w.]+)`)}
	for _, opt := range opts {
		opt(packer)
	}
	if packer.jarPath == "" {
		packer.jarPath = "./pack.jar"
	}
	if _, err := pathExists(packer.jarPath); err != nil {
		return nil, fmt.Errorf("jar path: %w, please copy the pack.jar to self project dir", err)
	}
	return packer, nil
}

type command string

func newCommand(fmtCmd string, args ...interface{}) command {
	return command(fmt.Sprintf(fmtCmd, args...))
}

func (c command) string() string {
	// 注入过滤
	s := string(c)
	for _, v := range []string{"|", "&", ";", "`", ">", "<"} {
		s = strings.ReplaceAll(s, v, "\\"+v)
	}
	return s
}

func (c command) baseApkPath(path string) {
}

func (c command) exec(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "java", strings.Split(c.string(), " ")...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(stdout)
	outs := make([]string, 0)
	for {
		output, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		outs = append(outs, string(output))
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
	}

	return strings.Join(outs, ""), err
}

func (w *Packer) rootCmd() command {
	cmd := "-jar %s"
	return newCommand(cmd, w.jarPath)
}

func (w *Packer) showCmd(apkPath string) command {
	cmd := "%s show %s"
	return newCommand(cmd, w.rootCmd(), apkPath)
}

func (w *Packer) putCmd(channels []string, baseApkPath, newApkPath string, extra map[string]string) command {
	var cmd string
	putCmd := "put -c"
	if len(channels) > 1 {
		putCmd = "batch -c"
	}
	channelsStr := strings.Join(channels, ",")
	// 指定渠道 channels 可以指定多个 用逗号分割 例: "toutiao1,toutiao2"
	cmd = fmt.Sprintf("%s %s %s %s", w.rootCmd(), putCmd, channelsStr, baseApkPath)
	// 指定生成的新apk包地址
	if newApkPath != "" {
		cmd = fmt.Sprintf("%s %s %s %s %s", w.rootCmd(), putCmd, channelsStr, baseApkPath, newApkPath)
	}
	// 添加额外字段信息
	if extra != nil {
		var extraArgs []string
		for k, v := range extra {
			extraArgs = append(extraArgs, fmt.Sprintf("%s=%s", k, v))
		}
		extraCmd := strings.Join(extraArgs, ",")
		cmd = fmt.Sprintf("%s %s %s -e %s %s %s", w.rootCmd(), putCmd, channelsStr, extraCmd, baseApkPath, newApkPath)
	}
	return command(cmd)
}

func (w *Packer) ShowApk(ctx context.Context, apkPath string) (map[string]string, error) {
	if strings.HasSuffix(apkPath, "*") {
		return nil, errors.New("just apk, can't be dir")
	}
	// 显示当前apk中的渠道和额外信息：
	output, err := w.showCmd(apkPath).exec(ctx)
	if err != nil {
		return nil, err
	}
	matched := w._extract.FindAll([]byte(output), -1)
	if len(matched) == 0 {
		return nil, fmt.Errorf("no meta info")
	}
	res := map[string]string{}
	for _, i := range matched {
		kv := strings.Split(string(i), "=")
		res[kv[0]] = kv[1]
	}
	return res, nil
}

func (w *Packer) ShowDir(ctx context.Context, dir string) (string, error) {
	if !strings.HasSuffix(dir, "*") {
		return "", errors.New("dir must end with [*]")
	}
	// 显示当前apk中的渠道和额外信息：
	output, err := w.showCmd(dir).exec(ctx)
	if err != nil {
		return "", err
	}
	return output, nil
}

func (w *Packer) PutChannel(ctx context.Context, channel string, baseApkPath, newApkPath string) (string, error) {
	// 写入渠道, 可指定输出新apk
	return w.putCmd([]string{channel}, baseApkPath, newApkPath, nil).exec(ctx)
}

func (w *Packer) PutChannelWithExtra(ctx context.Context, channel string, baseApkPath, newApkPath string, extra map[string]string) (string, error) {
	// 写入额外信息，不提供渠道时不写入渠道
	return w.putCmd([]string{channel}, baseApkPath, newApkPath, extra).exec(ctx)
}

func (w *Packer) PutExtra(ctx context.Context, baseApkPath, newApkPath string, extra map[string]string) (string, error) {
	// 写入额外信息，不提供渠道时不写入渠道
	return w.putCmd([]string{}, baseApkPath, newApkPath, extra).exec(ctx)
}

func (w *Packer) BatchChannels(ctx context.Context, channels []string, baseApkPath string) (string, error) {
	// 指定渠道列表, 自动生成新apk
	return w.putCmd(channels, baseApkPath, "", nil).exec(ctx)
}

func (w *Packer) BatchChannelsWithExtra(ctx context.Context, channels []string, baseApkPath string, extra map[string]string) (string, error) {
	// 指定渠道列表 携带额外信息
	return w.putCmd(channels, baseApkPath, "", extra).exec(ctx)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	return false, err
}
