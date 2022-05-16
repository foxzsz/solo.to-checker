package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	UsernameList []string
	ProxyList    []string
	Checked      int
	Available    int
	Unavailable  int
	errors       int
	Client       = http.Client{Timeout: 10 * time.Second}
)

func Check(Username string) {
RESTART:
	checkreq, err := http.NewRequest("GET", fmt.Sprintf("https://www.solo.to/%s", Username), nil)
	if err != nil {
		fmt.Println(err)
		errors++
	}

	proxyURL, err := url.Parse(fmt.Sprintf("http://%s", ProxyList[rand.Intn(len(ProxyList))]))
	if err != nil {
		fmt.Println(err)
		errors++
	}

	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	SoloResponse, err := Client.Do(checkreq)
	if err != nil {
		fmt.Println(fmt.Sprintf("error while doing req, retrying"))
		errors++
		goto RESTART
	}
	defer SoloResponse.Body.Close()

	switch SoloResponse.StatusCode {
	case 200:
		fmt.Println(fmt.Sprintf("username %s is unavailable", Username))
		Unavailable++
	case 204:
		fmt.Println(fmt.Sprintf("username %s is unavailable", Username))
		Unavailable++
	case 404:
		Available++
		availablefile, err := os.OpenFile("available.txt", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
		}
		availablefile.WriteString(Username + "\n")
		availablefile.Close()
		fmt.Println(fmt.Sprintf("username: %s is available!", Username))
	case 429:
		fmt.Println(fmt.Sprintf("ratelimited retrying in 5 seconds.."))
		time.Sleep(5 * time.Second)
		goto RESTART
	}

	Checked++

	cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf("Solo.to Checker by Dark | Checks: %d | Available: %d | Unavailable: %d | Errors: %d ", Checked, Available, Unavailable, errors))
	cmderr := cmd.Run()
	if cmderr != nil {
		fmt.Println(cmderr)
	}

}

func worker(wg *sync.WaitGroup, jobs <-chan string) {
	for j := range jobs {
		Check(j)
	}
	wg.Done()
}

//
func UserInput(m string) string {
	reader := bufio.NewReader(os.Stdin)
	var out string
	color.White(fmt.Sprintf("[input] %s : ", m))
	out, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	out = strings.TrimSuffix(out, "\r\n")
	out = strings.TrimSuffix(out, "\n")
	return out
}

func main() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	color.Green("Made By ogu.gg/Dark")
	color.Green("Add your proxies in proxies.txt and usernames in usernames.txt")

	proxyFile, err := os.Open("proxies.txt")
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(proxyFile)

	for scanner.Scan() {
		ProxyList = append(ProxyList, scanner.Text())
	}

	if len(ProxyList) == 0 {
		color.Red("Add your proxies in proxies.txt!")
		os.Exit(0)
	}

	usernameFile, err := os.Open("usernames.txt")
	if err != nil {
		fmt.Println(err)
	}

	scanner = bufio.NewScanner(usernameFile)

	for scanner.Scan() {
		UsernameList = append(UsernameList, scanner.Text())
	}

	if len(UsernameList) == 0 {
		color.Red("Add your usernames in usernames.txt!")
		os.Exit(0)
	}

	threadinput := UserInput("Enter amount of threads to check with, do not use too many")

	threads, err := strconv.Atoi(threadinput)
	if err != nil {
		// handle error
		color.Red("Invalid thread amount!")
		os.Exit(0)
	}

	wg := &sync.WaitGroup{}

	UsernameChannel := make(chan string)

	for i := 0; i <= threads; i++ {
		wg.Add(1)
		go worker(wg, UsernameChannel)
	}

	for _, a := range UsernameList {
		UsernameChannel <- a
	}
	close(UsernameChannel)
	wg.Wait()
	fmt.Println("Done checking")

}
