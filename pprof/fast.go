package main

import (
	// "encoding/json"
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	r := regexp.MustCompile("@")
	seenBrowsers := make(map[string]bool)
	uniqueBrowsers := 0
	foundUsers := ""
	i := -1
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		user := User{}
		_ = user.UnmarshalJSON([]byte(line))
		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				if _, in := seenBrowsers[browser]; !in {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}

			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				if _, in := seenBrowsers[browser]; !in {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
		}
		i++
		if !(isAndroid && isMSIE) {
			continue
		}
		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}
	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
