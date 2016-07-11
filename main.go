package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var input, output []string

func main() {
	cmd := os.Args[1]

	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	output = strings.Split(string(out), "\n")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input = append(input, scanner.Text())
	}

	for _, v := range input {
		found := false
		for _, w := range output {
			if v == w {
				found = true
			}
		}
		if !found {
			fmt.Println(v)
		}
	}
}
