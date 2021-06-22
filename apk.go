package walle

import (
    "errors"
    "path/filepath"
)

// apk the base apk, to generate new apk with channel or extras.
type apk struct {
    path string
    info channelInfo
}

func (a *apk) Path() string {
    return a.path
}

func (a *apk) Channel() string {
    return a.info.channel
}

func (a *apk) Extras() map[string]string {
    return a.info.extras
}

func (a *apk) All() map[string]string {
    res := map[string]string{
        "channel": a.Channel(),
    }
    for k, v := range a.Extras() {
        res[k] = v
    }
    return res
}

func NewApk(path string) (*apk, error) {
    if !isRegularFile(path) {
        return nil, errors.New("is not regular file")
    }
    info, err := readInfo(path)
    if err != nil {
        return nil, err
    }
    return &apk{
        path: path,
        info: info,
    }, nil
}

func (a *apk) PutChannel(ch, newPath string) (*apk, error) {
    outs, err := a.generate(newPath, []string{ch}, nil)
    if err != nil {
        return nil, err
    }
    return outs[0], err
}

func (a *apk) PutChannelWithExtra(ch string, extra map[string]string, newPath string) (*apk, error) {
    outs, err := a.generate(newPath, []string{ch}, extra)
    if err != nil {
        return nil, err
    }
    return outs[0], err
}

func (a *apk) PutExtra(newPath string, extra map[string]string) (*apk, error) {
    outs, err := a.generate(newPath, nil, extra)
    if err != nil {
        return nil, err
    }
    return outs[0], err
}

func (a *apk) BatchChannels(chs []string) ([]*apk, error) {
    outs, err := a.generate("", chs, nil)
    if err != nil {
        return nil, err
    }
    return outs, err
}

func (a *apk) BatchChannelsWithExtra(chs []string, extra map[string]string) ([]*apk, error) {
    outs, err := a.generate("", chs, extra)
    if err != nil {
        return nil, err
    }
    return outs, err
}

func (a *apk) generate(out string, channels []string, extras map[string]string) ([]*apk, error) {
    z, err := newZipSections(a.path)
    if err != nil {
        return nil, newErrf("Error occurred on parsing apk %s, %s", a.path, err)
    }
    if extras == nil {
        extras = a.Extras()
    }

    inputDir := filepath.Dir(a.path)

    name, ext := fileNameAndExt(a.path)
    outs := make([]*apk, len(channels))
    for i, channel := range channels {
        output := out
        if output == "" {
            output = filepath.Join(inputDir, name+"-"+channel+ext)
        }
        c := channelInfo{channel: channel, extras: extras}
        err = gen(c, z, output)
        if err != nil {
            return nil, newErrf("Error occurred on generating channel %s, %s", channel, err)
        }
        outs[i] = &apk{
            path: output,
            info: c,
        }
    }
    return outs, nil
}
