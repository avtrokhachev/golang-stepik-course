package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Только чтобы прокидывать в функцию префикс. Принимает параметры и прокидывает их во внутреннюю функцию.
// В golang нет аргументов по умолчанию, поэтому тесты бы не сработали, если я бы добавил аргумент
// в основную функцию
func dirTree(out io.Writer, path string, printFiles bool) error {
	err := dirTreeInner(out, path, printFiles, new(bytes.Buffer))
	return err
}

// Удаляет все файлы из среза директории, вызывается только если проставлен флаг
func filterFilesInDir(files []os.DirEntry) (filteredFilesInDir []os.DirEntry) {
	for _, file := range files {
		if file.IsDir() {
			filteredFilesInDir = append(filteredFilesInDir, file)
		}
	}
	return filteredFilesInDir
}

// Основная функция
func dirTreeInner(out io.Writer, path string, printFiles bool, buffer *bytes.Buffer) error {
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
		filePath = filePath + "───" + file.Name()

		if file.IsDir() {
			// Интересно, тут из-за append создаю дополнительный slice тупо чтобы прокинуть его в out.Write
			// мог бы просто в два чтения это сделать без создания нового "жирного" (с большим префиксом)
			// но решил не делать, чтобы не ухудшать и так грязный код
			_, err = out.Write(append(buffer.Bytes(), []byte(filePath+"\n")...))
			if err != nil {
				return err
			}

			var newPrefix string
			if i+1 != len(allFilesInDir) {
				newPrefix = "│"
			}
			newPrefix += "\t"

			// Возможно можно было бы сделать лучше, если использовать еще один флаг
			// "последний ли я или нет"
			// и высчитывать новый префекс прямо в начале и удалять через defer
			// тут такое не прокатит потому что префикс меняется ДО выхода из функции
			buffer.Write([]byte(newPrefix))
			err = dirTreeInner(out, path+"/"+file.Name(), printFiles, buffer)
			if err != nil {
				return err
			}
			buffer.Truncate(buffer.Len() - len(newPrefix))
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
			_, err = out.Write(append(buffer.Bytes(), []byte(filePath+"\n")...))
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
