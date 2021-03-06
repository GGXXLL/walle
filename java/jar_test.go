package wallepack

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	jar         *Packer
	testApkPath = `E:\GoProjects\src\ad_gains\packapk\base\b62a2d38360ea00854bff6df1d98161d.apk`
)

func TestNewPacker(t *testing.T) {
	o, _ := NewPacker()
	fmt.Println(o)
}

func TestPackApkJar_ShowApk(t *testing.T) {
	jar, _ = NewPacker()
	r, err := jar.ShowApk(context.Background(), testApkPath)
	assert.NoError(t, err)
	fmt.Println(r)
}

func TestPackApkJar_ShowDir(t *testing.T) {
	jar, _ = NewPacker()
	r, err := jar.ShowDir(context.Background(), "E:\\GoProjects\\src\\ad_gains\\packapk\\base\\*")
	assert.NoError(t, err)
	fmt.Println(r)
}

func TestPackApkJar_PutChannel(t *testing.T) {
	jar, _ = NewPacker()
	r, err := jar.PutChannel(context.Background(), "test", testApkPath, "")
	assert.NoError(t, err)
	fmt.Println(r)
}

func TestPackApkJar_BatchChannels(t *testing.T) {
	jar, _ = NewPacker()
	r, err := jar.BatchChannels(context.Background(), []string{"test", "test1", "test2"}, testApkPath)
	assert.NoError(t, err)
	fmt.Println(r)
}

func TestReplace(t *testing.T) {
	re, _ := regexp.Compile("([&|;`]+)")
	data := "echo 1 && echo 2 || echo 1; echo 12; echo 1 & echo3 & `echo 1`"
	data = re.ReplaceAllString(data, "\\/1")
	fmt.Println(data)
}
