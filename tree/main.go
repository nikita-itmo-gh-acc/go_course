package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"slices"

	"github.com/idsulik/go-collections/stack"
)

type stackElem struct {
	Value  fs.DirEntry
	IsLast bool
	Depth  int
	Path   string
	Base   string
}

const (
	beforeRegular string = "├───"
	beforeLast    string = "└───"
	baseLayer     string = "│   "
	layer         string = "    "
)

func renderBase(lastBase string, isInEnd bool) (res string) {
	res = lastBase
	if isInEnd {
		res += layer
	} else {
		res += baseLayer
	}
	return
}

func renderRow(base string, isLast bool, elemName string) (res string) {
	res = base
	if isLast {
		res += beforeLast
	} else {
		res += beforeRegular
	}
	res += elemName + "\n"
	return
}

func pushMany(list *[]fs.DirEntry, stack *stack.Stack[stackElem], depth int, dir_path string, base string) {
	for i, content := range slices.Backward(*list) {
		var isLast bool
		if i == len(*list)-1 {
			isLast = true
		}
		(*stack).Push(stackElem{Value: content, IsLast: isLast, Depth: depth, Path: dir_path + "\\" + content.Name(), Base: base})
	}
}

func dirTree(output io.Writer, path string, wFiles bool) error {
	curDir, _ := os.Getwd()
	curDir += "\\" + path
	if err := os.Chdir(curDir); err != nil {
		return errors.New("can't change directory")
	}
	content_stack := stack.New[stackElem](128)
	contents, _ := os.ReadDir(curDir)
	pushMany(&contents, content_stack, 0, curDir, "")
	for {
		if content_stack.IsEmpty() {
			return nil
		}
		elem, _ := content_stack.Pop()
		output.Write([]byte(renderRow(elem.Base, elem.IsLast, elem.Value.Name())))
		if elem.Value.IsDir() {
			dir_contents, _ := os.ReadDir(elem.Path)
			pushMany(&dir_contents, content_stack, elem.Depth+1, elem.Path, renderBase(elem.Base, elem.IsLast))
		}
	}
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
