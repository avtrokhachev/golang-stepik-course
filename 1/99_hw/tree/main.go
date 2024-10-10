package main

import (
	"fmt"
	"io"
	"os"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := dirTreeInner(out, path, printFiles, "")
	return err
}

func filterFilesInDir(files []os.DirEntry) (filteredFilesInDir []os.DirEntry) {
	for _, file := range files {
		if file.IsDir() {
			filteredFilesInDir = append(filteredFilesInDir, file)
		}
	}
	return filteredFilesInDir
}

func dirTreeInner(out io.Writer, path string, printFiles bool, prefix string) error {
	allFilesInDir, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	if !printFiles {
		allFilesInDir = filterFilesInDir(allFilesInDir)
	}

	for i, file := range allFilesInDir {
		var filePath = "├"
		if i+1 == len(allFilesInDir) {
			filePath = "└"
		}
		filePath = prefix + filePath + "───" + file.Name()

		if file.IsDir() {
			_, err = out.Write([]byte(filePath + "\n"))
			if err != nil {
				return err
			}

			var newPrefix string
			if i+1 != len(allFilesInDir) {
				newPrefix = "│"
			}
			newPrefix += "\t"

			err = dirTreeInner(out, path+"/"+file.Name(), printFiles, prefix+newPrefix)
			if err != nil {
				return err
			}
		} else if printFiles {
			fileInfo, err := file.Info()
			if err != nil {
				return err
			}

			outputSize := "empty"
			if size := fileInfo.Size(); size > 0 {
				outputSize = fmt.Sprintf("%db", size)
			}
			filePath = filePath + fmt.Sprintf(" (%s)", outputSize)
			_, err = out.Write([]byte(filePath + "\n"))
			if err != nil {
				return err
			}
		}
	}
	return nil
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
