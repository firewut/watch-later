package args

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var (
	Flags = map[string]string{}
)

func init() {
	Flags["mode"] = ""

	// If we passed args thru args
	if len(os.Args[1:]) > 0 {
		for _, arg := range os.Args[1:] {
			parts := strings.SplitN(strings.TrimPrefix(arg, "--"), "=", 2)
			key := strings.ToLower(strings.Trim(parts[0], " \t\n"))
			if len(parts) > 1 {
				value := strings.Trim(parts[1], " \t\n")
				Flags[key] = value
			}
		}
	}

	// If we passed args thru os env
	for key, _ := range Flags {
		v := os.Getenv(key)
		if len(v) > 0 {
			Flags[key] = v
		}
	}

	reader := bufio.NewReader(os.Stdin)
	defer os.Stdin.Close()
	if _, exists := Flags["input"]; exists {
		for {
			line, _, _ := reader.ReadLine()
			parts := strings.SplitN(string(line), "=", 2)
			key := strings.ToLower(strings.Trim(parts[0], " \t\n"))
			if key == "" {
				break
			}
			value := strings.Trim(parts[1], " \t\n")
			_, exists := Flags[key]
			if exists {
				Flags[key] = value
				log.Println("Set value for ", key, value, Flags[key])
			} else {
				log.Println("Not found key ", key)
			}
		}
	}
}
