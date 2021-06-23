package _go

import (
	"fmt"
	"os"
)

type zipSections struct {
	beforeSigningBlock []byte
	signingBlock       []byte
	signingBlockOffset int64
	centraDir          []byte
	centralDirOffset   int64
	eocd               []byte
	eocdOffset         int64
}
type transform func(*zipSections) (*zipSections, error)

func (z *zipSections) writeTo(output string, transform transform) (err error) {
	f, err := os.Create(output)
	if err != nil {
		return
	}

	defer f.Close()

	newZip, err := transform(z)
	if err != nil {
		return
	}

	for _, s := range [][]byte{
		newZip.beforeSigningBlock,
		newZip.signingBlock,
		newZip.centraDir,
		newZip.eocd} {
		_, err := f.Write(s)
		if err != nil {
			return err
		}
	}
	return
}

func newZipSections(input string) (z zipSections, err error) {
	in, err := os.Open(input)
	if err != nil {
		return
	}
	defer in.Close()

	// read eocd
	eocd, eocdOffset, err := findEndOfCentralDirectoryRecord(in)
	if err != nil {
		return
	}
	centralDirOffset := getEocdCentralDirectoryOffset(eocd)
	centralDirSize := getEocdCentralDirectorySize(eocd)
	z.eocd = eocd
	z.eocdOffset = eocdOffset
	z.centralDirOffset = int64(centralDirOffset)

	// read signing block
	signingBlock, signingBlockOffset, err := findApkSigningBlock(in, centralDirOffset)
	if err != nil {
		return
	}
	z.signingBlock = signingBlock
	z.signingBlockOffset = signingBlockOffset
	// read bytes before signing block
	//TODO: waste too large memory
	if signingBlockOffset >= 64*1024*1024 {
		fmt.Print("Warning: maybe waste large memory on processing this apk! ")
		fmt.Println("Before APK Signing Block bytes size is", signingBlockOffset/1024/1024, "MB")
	}
	beforeSigningBlock := make([]byte, signingBlockOffset)
	n, err := in.ReadAt(beforeSigningBlock, 0)
	if err != nil {
		return
	}
	if int64(n) != signingBlockOffset {
		return z, fmt.Errorf("Read bytes count mismatched! Expect %d, but %d", signingBlockOffset, n)
	}
	z.beforeSigningBlock = beforeSigningBlock

	centralDir := make([]byte, centralDirSize)
	n, err = in.ReadAt(centralDir, int64(centralDirOffset))
	if uint32(n) != centralDirSize {
		return z, fmt.Errorf("Read bytes count mismatched! Expect %d, but %d", centralDirSize, n)
	}
	z.centraDir = centralDir
	return
}

func gen(info channelInfo, sections zipSections, output string) error {
	_, err := os.Stat(output)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return sections.writeTo(output, newTransform(info))
}

func newTransform(info channelInfo) transform {
	return func(zip *zipSections) (*zipSections, error) {

		newBlock, diffSize, err := makeSigningBlockWithInfo(info, zip.signingBlock)
		if err != nil {
			return nil, err
		}
		newzip := new(zipSections)
		newzip.beforeSigningBlock = zip.beforeSigningBlock
		newzip.signingBlock = newBlock
		newzip.signingBlockOffset = zip.signingBlockOffset
		newzip.centraDir = zip.centraDir
		newzip.centralDirOffset = zip.centralDirOffset
		newzip.eocdOffset = zip.eocdOffset
		newzip.eocd = makeEocd(zip.eocd, uint32(int64(diffSize)+zip.centralDirOffset))
		return newzip, nil
	}
}

func newErrf(f string, args ...interface{}) error {
	return fmt.Errorf(f, args...)
}
