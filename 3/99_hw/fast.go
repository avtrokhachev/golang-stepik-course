package main

import (
	"bufio"
	"fmt"
	"github.com/mailru/easyjson"
	"hw3/json"
	"io"
	"os"
	"regexp"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Fprint(out, "found users:\n")

	regexpMSIE := regexp.MustCompile("MSIE")
	regexpAndroid := regexp.MustCompile("Android")
	seenBrowsers := make(map[string]bool)
	uniqueBrowsers := 0
	outputTemplate := "[%d] %s <%s>\n"

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	i := -1
	for fileScanner.Scan() {
		i++
		user := &browserusers.User{}
		err := easyjson.Unmarshal(fileScanner.Bytes(), user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for ind := range user.Browsers {
			var ok bool
			if ok = regexpAndroid.MatchString(user.Browsers[ind]); ok {
				isAndroid = true
				_, notSeenBefore := seenBrowsers[user.Browsers[ind]]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[user.Browsers[ind]] = true
					uniqueBrowsers++
				}
			}
			if ok = regexpMSIE.MatchString(user.Browsers[ind]); ok {
				isMSIE = true
				_, notSeenBefore := seenBrowsers[user.Browsers[ind]]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[user.Browsers[ind]] = true
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		fmt.Fprintf(out, outputTemplate, i, user.Name, email)
	}

	fmt.Fprint(out, "\n")
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
