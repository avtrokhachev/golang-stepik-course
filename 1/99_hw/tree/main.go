package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Не придумал названия лучше :/
type ObjectInDir interface {
	IsDir() bool
	fmt.Stringer
}

type File struct {
	Name string
	Size int64
}

func (f *File) IsDir() bool {
	return false
}

func (f *File) String() string {
	sizeString := "empty"
	if size := f.Size; size > 0 {
		sizeString = fmt.Sprintf("%db", size)
	}

	return fmt.Sprintf("%s (%s)", f.Name, sizeString)
}

type Directory struct {
	Name string
}

func (d *Directory) IsDir() bool {
	return true
}

func (d *Directory) String() string {
	return d.Name
}

func ConvertDirEntryToObjectInDir(dirEntry *os.DirEntry) (ObjectInDir, error) {
	if (*dirEntry).IsDir() {
		return &Directory{
			Name: (*dirEntry).Name(),
		}, nil
	}

	fileInfo, err := (*dirEntry).Info()
	if err != nil {
		return nil, err

	}

	return &File{
		Name: (*dirEntry).Name(),
		Size: fileInfo.Size(),
	}, nil
}

func getAllObjectsInDir(path string, filterFiles bool) ([]ObjectInDir, error) {
	allDirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if !filterFiles {
		allDirEntries = filterFilesInDir(allDirEntries)
	}

	objectsInDir := make([]ObjectInDir, 0)
	for _, dirEntry := range allDirEntries {
		object, err := ConvertDirEntryToObjectInDir(&dirEntry)
		if err != nil {
			return nil, err
		}

		objectsInDir = append(objectsInDir, object)
	}

	return objectsInDir, nil
}

func filterFilesInDir(files []os.DirEntry) (filteredFilesInDir []os.DirEntry) {
	for _, file := range files {
		if file.IsDir() {
			filteredFilesInDir = append(filteredFilesInDir, file)
		}
	}
	return filteredFilesInDir
}

func dirTreeInner(out io.Writer, path string, printFiles bool, buffer *bytes.Buffer) error {
	allObjectsInDir, err := getAllObjectsInDir(path, printFiles)
	if err != nil {
		return err
	}

	getFilePathPrefix := func(isLast bool) string {
		if isLast {
			return "└"
		}
		return "├"
	}
	getInnerDirectoryPrefix := func(isLast bool) string {
		prefix := ""
		if !isLast {
			prefix = "│"
		}
		return prefix + "\t"
	}

	for i, file := range allObjectsInDir {
		isLast := (i + 1) == len(allObjectsInDir)
		filePath := getFilePathPrefix(isLast) + "───" + file.String()

		if file.IsDir() {
			_, err = out.Write(append(buffer.Bytes(), []byte(filePath+"\n")...))
			if err != nil {
				return err
			}

			// Возможно можно было бы сделать лучше, если использовать еще один флаг
			// "последний ли я или нет"
			// и высчитывать новый префекс прямо в начале и удалять через defer
			// тут такое не прокатит потому что префикс меняется ДО выхода из функции
			newPrefix := getInnerDirectoryPrefix(isLast)
			buffer.Write([]byte(newPrefix))
			err = dirTreeInner(out, path+"/"+file.(*Directory).Name, printFiles, buffer)
			if err != nil {
				return err
			}
			buffer.Truncate(buffer.Len() - len(newPrefix))
		} else if printFiles {
			_, err = out.Write(append(buffer.Bytes(), []byte(filePath+"\n")...))
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dirTreeInner(out, path, printFiles, new(bytes.Buffer))
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
