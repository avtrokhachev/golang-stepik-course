Для начала скопировал код из common.go в fast.go чтобы постепенно его улучшать и закрывать узкие места.

Снял профиль по памяти
```
avtrokhachev@i106040781 99_hw % go test -bench . -benchmem   
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8               20          59688473 ns/op        20374506 B/op     182855 allocs/op
BenchmarkFast-8               18          64583444 ns/op        20370582 B/op     182856 allocs/op
PASS
ok      hw3     3.840s
```

Вижу огромное кол-во 182856 allocs/op и 20370582 B/op что говорит о лишнем выделении памяти. Попытаюсь соптимизировать, поняв где она выделяется.

```
(pprof) top
Showing nodes accounting for 73.40MB, 91.69% of 80.05MB total
Dropped 110 nodes (cum <= 0.40MB)
Showing top 10 nodes out of 48
      flat  flat%   sum%        cum   cum%
   29.31MB 36.61% 36.61%    29.31MB 36.61%  regexp/syntax.(*compiler).inst (inline)
   13.97MB 17.45% 54.06%    13.97MB 17.45%  io.ReadAll
   10.26MB 12.81% 66.87%    10.26MB 12.81%  regexp/syntax.(*parser).newRegexp (inline)
    5.09MB  6.36% 73.23%    52.51MB 65.59%  regexp.compile
    3.91MB  4.88% 78.11%    15.87MB 19.83%  regexp/syntax.parse
    3.29MB  4.12% 82.23%    40.76MB 50.92%  hw3.SlowSearch
    2.75MB  3.43% 85.66%    39.27MB 49.06%  hw3.FastSearch
    1.76MB  2.20% 87.86%     1.76MB  2.20%  regexp.(*bitState).reset
    1.60MB  2.00% 89.85%     1.60MB  2.00%  encoding/json.unquote (inline)
    1.47MB  1.83% 91.69%     2.93MB  3.66%  regexp/syntax.(*compiler).init (inline)
    
(pprof) list FastSearch
Total: 80.05MB
ROUTINE ======================== hw3.FastSearch in /Users/avtrokhachev/Downloads/golang_web_services_2024-04-26/3/99_hw/fast.go
    2.75MB    39.27MB (flat, cum) 49.06% of Total
         .          .     14:func FastSearch(out io.Writer) {
         .          .     15:   /*
         .          .     16:           !!! !!! !!!
         .          .     17:           обратите внимание - в задании обязательно нужен отчет
         .          .     18:           делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     19:           так же обратите внимание на команду в параметром -http
         .          .     20:           перечитайте еще раз задание
         .          .     21:           !!! !!! !!!
         .          .     22:   */
         .       256B     23:   file, err := os.Open(filePath)
         .          .     24:   if err != nil {
         .          .     25:           panic(err)
         .          .     26:   }
         .          .     27:
         .     6.74MB     28:   fileContents, err := ioutil.ReadAll(file)
         .          .     29:   if err != nil {
         .          .     30:           panic(err)
         .          .     31:   }
         .          .     32:
         .     1.52kB     33:   r := regexp.MustCompile("@")
         .          .     34:   seenBrowsers := []string{}
         .          .     35:   uniqueBrowsers := 0
         .          .     36:   foundUsers := ""
         .          .     37:
    1.09MB     1.12MB     38:   lines := strings.Split(string(fileContents), "\n")
         .          .     39:
         .          .     40:   users := make([]map[string]interface{}, 0)
         .          .     41:   for _, line := range lines {
  109.38kB   109.38kB     42:           user := make(map[string]interface{})
         .          .     43:           // fmt.Printf("%v %v\n", err, line)
    1.13MB     3.89MB     44:           err := json.Unmarshal([]byte(line), &user)
         .          .     45:           if err != nil {
         .          .     46:                   panic(err)
         .          .     47:           }
   34.23kB    34.23kB     48:           users = append(users, user)
         .          .     49:   }
         .          .     50:
         .          .     51:   for i, user := range users {
         .          .     52:
         .          .     53:           isAndroid := false
         .          .     54:           isMSIE := false
         .          .     55:
         .          .     56:           browsers, ok := user["browsers"].([]interface{})
         .          .     57:           if !ok {
         .          .     58:                   // log.Println("cant cast browsers")
         .          .     59:                   continue
         .          .     60:           }
         .          .     61:
         .          .     62:           for _, browserRaw := range browsers {
         .          .     63:                   browser, ok := browserRaw.(string)
         .          .     64:                   if !ok {
         .          .     65:                           // log.Println("cant cast browser to string")
         .          .     66:                           continue
         .          .     67:                   }
         .    16.40MB     68:                   if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
         .          .     69:                           isAndroid = true
         .          .     70:                           notSeenBefore := true
         .          .     71:                           for _, item := range seenBrowsers {
         .          .     72:                                   if item == browser {
         .          .     73:                                           notSeenBefore = false
         .          .     74:                                   }
         .          .     75:                           }
         .          .     76:                           if notSeenBefore {
         .          .     77:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
    5.59kB     5.59kB     78:                                   seenBrowsers = append(seenBrowsers, browser)
         .          .     79:                                   uniqueBrowsers++
         .          .     80:                           }
         .          .     81:                   }
         .          .     82:           }
         .          .     83:
         .          .     84:           for _, browserRaw := range browsers {
         .          .     85:                   browser, ok := browserRaw.(string)
         .          .     86:                   if !ok {
         .          .     87:                           // log.Println("cant cast browser to string")
         .          .     88:                           continue
         .          .     89:                   }
         .    10.53MB     90:                   if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
         .          .     91:                           isMSIE = true
         .          .     92:                           notSeenBefore := true
         .          .     93:                           for _, item := range seenBrowsers {
         .          .     94:                                   if item == browser {
         .          .     95:                                           notSeenBefore = false
         .          .     96:                                   }
         .          .     97:                           }
         .          .     98:                           if notSeenBefore {
         .          .     99:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
    3.12kB     3.12kB    100:                                   seenBrowsers = append(seenBrowsers, browser)
         .          .    101:                                   uniqueBrowsers++
         .          .    102:                           }
         .          .    103:                   }
         .          .    104:           }
         .          .    105:
         .          .    106:           if !(isAndroid && isMSIE) {
         .          .    107:                   continue
         .          .    108:           }
         .          .    109:
         .          .    110:           // log.Println("Android and MSIE user:", user["name"], user["email"])
         .    17.48kB    111:           email := r.ReplaceAllString(user["email"].(string), " [at] ")
  371.22kB   403.02kB    112:           foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
         .          .    113:   }
         .          .    114:
    9.53kB    23.78kB    115:   fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    116:   fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    117:}
         .          .    118:
         .          .    119://func main() {
         .          .    120:// fmt.Println("Unused main function")

```

Как видно, 6.74MB выделяет при чтении всего файла, еще 1.12MB при разбивке на линии.
Unmarshall ест тоже не мало 3.89MB.
16.40MB и 10.53MB выделяет при сопоставлении регулярных выражений.

Поправил, посмотрю на бенчмарки теперь.

```
avtrokhachev@i106040781 99_hw % go test -bench . -benchmem -memprofile=mem.out -memprofilerate=1
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8                1        1747589689 ns/op        20499472 B/op     182870 allocs/op
BenchmarkFast-8               14          77302236 ns/op          963127 B/op      12825 allocs/op
```

Как видно теперь разница в 10+ раз, т.е. значительное улучшение. Посмотрю поподробнее.
```
ROUTINE ======================== hw3.FastSearch in /Users/avtrokhachev/Downloads/golang_web_services_2024-04-26/3/99_hw/fast.go
   11.56MB    32.43MB (flat, cum) 44.50% of Total
         .          .     14:func FastSearch(out io.Writer) {
         .          .     15:   /*
         .          .     16:           !!! !!! !!!
         .          .     17:           обратите внимание - в задании обязательно нужен отчет
         .          .     18:           делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     19:           так же обратите внимание на команду в параметром -http
         .          .     20:           перечитайте еще раз задание
         .          .     21:           !!! !!! !!!
         .          .     22:   */
         .     4.38kB     23:   file, err := os.Open(filePath)
         .          .     24:   if err != nil {
         .          .     25:           panic(err)
         .          .     26:   }
         .          .     27:   defer func() {
         .          .     28:           err = file.Close()
         .          .     29:           if err != nil {
         .          .     30:                   panic(err)
         .          .     31:           }
         .          .     32:   }()
         .          .     33:
         .    26.52kB     34:   r := regexp.MustCompile("@")
         .          .     35:   seenBrowsers := []string{}
         .          .     36:   uniqueBrowsers := 0
         .          .     37:   foundUsers := ""
         .          .     38:
         .          .     39:   fileScanner := bufio.NewScanner(file)
         .          .     40:   fileScanner.Split(bufio.ScanLines)
         .          .     41:   users := make([]*browserusers.User, 0)
         .      140kB     42:   for fileScanner.Scan() {
    4.27MB     4.27MB     43:           user := &browserusers.User{}
         .          .     44:           // fmt.Printf("%v %v\n", err, line)
         .    19.42MB     45:           err := easyjson.Unmarshal(fileScanner.Bytes(), user)
         .          .     46:           if err != nil {
         .          .     47:                   panic(err)
         .          .     48:           }
  599.10kB   599.10kB     49:           users = append(users, user)
         .          .     50:   }
         .          .     51:
         .    46.21kB     52:   regexpMSIE := regexp.MustCompile("MSIE")
         .    71.37kB     53:   regexpAndroid := regexp.MustCompile("Android")
         .          .     54:   for i, user := range users {
         .          .     55:
         .          .     56:           isAndroid := false
         .          .     57:           isMSIE := false
         .          .     58:
         .          .     59:           for _, browser := range user.Browsers {
         .   594.62kB     60:                   if ok := regexpAndroid.MatchString(browser); ok {
         .          .     61:                           isAndroid = true
         .          .     62:                           notSeenBefore := true
         .          .     63:                           for _, item := range seenBrowsers {
         .          .     64:                                   if item == browser {
         .          .     65:                                           notSeenBefore = false
         .          .     66:                                   }
         .          .     67:                           }
         .          .     68:                           if notSeenBefore {
         .          .     69:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
   97.89kB    97.89kB     70:                                   seenBrowsers = append(seenBrowsers, browser)
         .          .     71:                                   uniqueBrowsers++
         .          .     72:                           }
         .          .     73:                   }
         .          .     74:           }
         .          .     75:
         .          .     76:           for _, browser := range user.Browsers {
         .          .     77:                   if ok := regexpMSIE.MatchString(browser); ok {
         .          .     78:                           isMSIE = true
         .          .     79:                           notSeenBefore := true
         .          .     80:                           for _, item := range seenBrowsers {
         .          .     81:                                   if item == browser {
         .          .     82:                                           notSeenBefore = false
         .          .     83:                                   }
         .          .     84:                           }
         .          .     85:                           if notSeenBefore {
         .          .     86:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
   54.69kB    54.69kB     87:                                   seenBrowsers = append(seenBrowsers, browser)
         .          .     88:                                   uniqueBrowsers++
         .          .     89:                           }
         .          .     90:                   }
         .          .     91:           }
         .          .     92:
         .          .     93:           if !(isAndroid && isMSIE) {
         .          .     94:                   continue
         .          .     95:           }
         .          .     96:
         .          .     97:           // log.Println("Android and MSIE user:", user["name"], user["email"])
         .   329.55kB     98:           email := r.ReplaceAllString(user.Email, " [at] ")
    6.39MB     6.57MB     99:           foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
         .          .    100:   }
         .          .    101:
  166.80kB   247.55kB    102:   fmt.Fprintln(out, "found users:\n"+foundUsers)
         .          .    103:   fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    104:}
```

На данный момент главная пробема это накопление данных, мы зачем-то собираем все в один slice вместо потоковой обработки.
Поправлю это.

```
avtrokhachev@i106040781 99_hw % go test -bench . -benchmem -memprofile=mem.out -memprofilerate=1
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8                1        1846186339 ns/op        20574048 B/op     182888 allocs/op
BenchmarkFast-8               18          69384240 ns/op          619827 B/op      11647 allocs/op
```

Результаты еще улучшились. Теперь экстремально соптимизирую накопление результата, буду хранить не сами строки а []byte (потому что под ними лижит всего 3 инта).
Для получения []byte из string без доп. памяти использую unsafe. Так же перепишу ненужную замену с регуляркой в конце.
Так же удалил лишние поля из json.

```
avtrokhachev@i106040781 99_hw % go test -bench . -benchmem -memprofile=mem.out -memprofilerate=1
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8                1        1809173177 ns/op        20536040 B/op     182867 allocs/op
BenchmarkFast-8               20          58657398 ns/op          556939 B/op       7352 allocs/op

  113.62kB    11.71MB (flat, cum) 21.99% of Total
         .          .     16:func FastSearch(out io.Writer) {
         .          .     17:   /*
         .          .     18:           !!! !!! !!!
         .          .     19:           обратите внимание - в задании обязательно нужен отчет
         .          .     20:           делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
         .          .     21:           так же обратите внимание на команду в параметром -http
         .          .     22:           перечитайте еще раз задание
         .          .     23:           !!! !!! !!!
         .          .     24:   */
         .     2.75kB     25:   file, err := os.Open(filePath)
         .          .     26:   if err != nil {
         .          .     27:           panic(err)
         .          .     28:   }
         .          .     29:   defer func() {
         .          .     30:           err = file.Close()
         .          .     31:           if err != nil {
         .          .     32:                   panic(err)
         .          .     33:           }
         .          .     34:   }()
         .          .     35:
         .     2.77kB     36:   fmt.Fprint(out, "found users:\n")
         .          .     37:
         .    29.05kB     38:   regexpMSIE := regexp.MustCompile("MSIE")
         .    44.86kB     39:   regexpAndroid := regexp.MustCompile("Android")
         .          .     40:   seenBrowsers := []*string{}
         .          .     41:   uniqueBrowsers := 0
         .          .     42:   outputTemplate := "[%d] %s <%s>\n"
         .          .     43:
         .          .     44:   fileScanner := bufio.NewScanner(file)
         .          .     45:   fileScanner.Split(bufio.ScanLines)
         .          .     46:   i := -1
         .       88kB     47:   for fileScanner.Scan() {
         .          .     48:           i++
         .          .     49:           user := &browserusers.User{}
         .    11.03MB     50:           err := easyjson.Unmarshal(fileScanner.Bytes(), user)
         .          .     51:           if err != nil {
         .          .     52:                   panic(err)
         .          .     53:           }
         .          .     54:
         .          .     55:           isAndroid := false
         .          .     56:           isMSIE := false
         .          .     57:
         .          .     58:           for ind := range user.Browsers {
         .          .     59:                   var browserBytes []byte = unsafe.Slice(unsafe.StringData(user.Browsers[ind]), len(user.Browsers[ind]))
         .          .     60:                   var ok bool
         .   331.23kB     61:                   if ok = regexpAndroid.Match(browserBytes); ok {
         .          .     62:                           isAndroid = true
         .          .     63:                           notSeenBefore := true
         .          .     64:                           for _, item := range seenBrowsers {
         .          .     65:                                   if *item == user.Browsers[ind] {
         .          .     66:                                           notSeenBefore = false
         .          .     67:                                   }
         .          .     68:                           }
         .          .     69:                           if notSeenBefore {
         .          .     70:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
   30.77kB    30.77kB     71:                                   seenBrowsers = append(seenBrowsers, &user.Browsers[ind])
         .          .     72:                                   uniqueBrowsers++
         .          .     73:                           }
         .          .     74:                   }
         .          .     75:                   if ok = regexpMSIE.Match(browserBytes); ok {
         .          .     76:                           isMSIE = true
         .          .     77:                           notSeenBefore := true
         .          .     78:                           for _, item := range seenBrowsers {
         .          .     79:                                   if *item == user.Browsers[ind] {
         .          .     80:                                           notSeenBefore = false
         .          .     81:                                   }
         .          .     82:                           }
         .          .     83:                           if notSeenBefore {
         .          .     84:                                   // log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
   15.81kB    15.81kB     85:                                   seenBrowsers = append(seenBrowsers, &user.Browsers[ind])
         .          .     86:                                   uniqueBrowsers++
         .          .     87:                           }
         .          .     88:                   }
         .          .     89:           }
         .          .     90:
         .          .     91:           if !(isAndroid && isMSIE) {
         .          .     92:                   continue
         .          .     93:           }
         .          .     94:
         .          .     95:           // log.Println("Android and MSIE user:", user["name"], user["email"])
         .    57.06kB     96:           email := strings.ReplaceAll(user.Email, "@", " [at] ")
   67.05kB    89.08kB     97:           fmt.Fprintf(out, outputTemplate, i, user.Name, email)
         .          .     98:   }
         .          .     99:
         .          .    100:   fmt.Fprint(out, "\n")
         .          .    101:   fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
         .          .    102:}

```

Теперь соптимизирую память, для этого посмотрю текущее положение:

```
avtrokhachev@i106040781 99_hw % go test -bench . -cpuprofile=cpu.out          
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8               25          48376910 ns/op
BenchmarkFast-8              213           5199205 ns/op
```

```
     10ms       10ms     65:                                   if *item == user.Browsers[ind] {
```

Заоптимайжу поиск в списке браузеров.

```
avtrokhachev@i106040781 99_hw % go test -bench . -benchmem -memprofile=mem.out -memprofilerate=1
goos: darwin
goarch: amd64
pkg: hw3
cpu: Intel(R) Core(TM) i5-1038NG7 CPU @ 2.00GHz
BenchmarkSlow-8                1        1729835433 ns/op        20424072 B/op     182846 allocs/op
BenchmarkFast-8               20          55921964 ns/op          567836 B/op       7356 allocs/op
```

Прошло с учетом того, что slow у меня отстает от тестового ~10 раз
