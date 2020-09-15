package vfs

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// File represents objects in the filesytem, is essentially an os.FileInfo with shadow details
// injected in
type FileInfo struct {
	os.FileInfo
	Owner string
	Group string
}

type FileList []FileInfo

// Short returns a string that lists the collection of files by name only,
// one per line
func (flist FileList) Short() []byte {
	var buf bytes.Buffer
	for _, file := range flist {
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	return buf.Bytes()
}

// Detailed returns a string that lists the collection of files with extra
// detail, one per line
func (flist FileList) Detailed() []byte {
	var buf bytes.Buffer
	for _, file := range flist {
		fmt.Fprint(&buf, file.Mode().String())
		fmt.Fprintf(&buf, " 1 %s %s ", file.Owner, file.Group)
		fmt.Fprint(&buf, lpad(strconv.FormatInt(file.Size(), 10), 12))
		fmt.Fprint(&buf, file.ModTime().Format(" Jan _2 15:04 "))
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	return buf.Bytes()
}

func (flist FileList) SortByName() {
	sort.Slice(flist, func(i, j int) bool {
		return flist[i].Name() < flist[j].Name()
	})
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length-len(input)) + input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}
