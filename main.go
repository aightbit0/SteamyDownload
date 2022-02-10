package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Refresh        int    `json:"refresh"`
	Path           string `json:"path"`
	TimeToShutdown string `json:"timetoshutdown"`
}

func main() {
	currentTime := time.Now()
	fmt.Println("start listening... ", currentTime.Format("2006-01-02 15:04:05"))
	fmt.Println("type in e for stop Routines")
	fmt.Println("type in a for stop shutdown (after game installed)")
	conf := loadJSONConfig("steamchecker.json")
	//time to refresh in seconds
	ticker := time.NewTicker(time.Duration(conf.Refresh) * time.Second)
	quit := make(chan struct{})

	//starts routine reading file
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("info: reading file...")
				shut := ReadLogFile(currentTime, *conf)
				if shut {
					close(quit)
				}
			case <-quit:
				ticker.Stop()
				fmt.Println("stop")
				return
			}
		}
	}()

	for {
		fmt.Print("action: -> ")
		scanner1 := bufio.NewScanner(os.Stdin)
		var typ string
		if scanner1.Scan() {
			typ = scanner1.Text()
		}

		if typ == "e" {
			close(quit)
		}

		if typ == "a" {
			killShutdown()
		}
	}

}

func ReadLogFile(currentTime time.Time, config Config) bool {

	file, err := os.Open(config.Path)

	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {

		//check if log exists with given parameters
		if strings.Contains(scanner.Text(), currentTime.Format("2006-01-02")) {
			if strings.Contains(scanner.Text(), "scheduler finished : removed from schedule (result No Error, state 0xc)") {
				s := strings.Split(scanner.Text(), "]")
				theDate := s[0]

				//parse target datetime
				t, err := time.Parse("2006-01-02 15:04:05", theDate[1:])
				if err != nil {
					log.Fatal(err)
				}

				//parse actual datetime
				t2, err := time.Parse("2006-01-02 15:04:05", currentTime.Format("2006-01-02 15:04:05"))
				if err != nil {
					log.Fatal(err)
				}

				if t.After(t2) {
					fmt.Println("game installed successfull !")
					fmt.Println("starting shutdown.....")
					fmt.Println("type in a to cancel shutdown")
					file.Close()
					shutdown(config.TimeToShutdown)
					return true
				}
			}
		}
	}

	file.Close()

	return false
}

func shutdown(tvalue string) {
	fmt.Println(tvalue)
	app := "shutdown"
	arg1 := "-s"
	arg2 := "-t"
	arg3 := tvalue
	cmd := exec.Command(app, arg1, arg2, arg3)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error(), stdout)
		return
	}
	return
}

func killShutdown() {
	cmd := exec.Command("shutdown", "-a")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error(), stdout)
		return
	}
	return
}

func loadJSONConfig(p string) *Config {
	data, err := os.Open(p)
	if err != nil {
		return nil
	}
	d := json.NewDecoder(data)
	var c Config
	if err := d.Decode(&c); err != nil {
		return nil
	}
	fmt.Println(c.Path)
	return &c
}
