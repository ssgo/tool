package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		printUsage()
		return
	}

	op := os.Args[1]
	switch op {
	case "-v":
		outs, _ := runCommand("git", "tag", "-l")
		for _, line := range outs {
			fmt.Println(line)
		}
		fmt.Println(len(outs), "Versions")
	case "-u":
		oldVer := "v0.0.0"
		outs, _ := runCommand("git", "tag", "-l")
		for i := len(outs) - 1; i >= 0; i-- {
			if outs[i][0] == 'v' && strings.IndexByte(outs[i], '.') != -1 {
				oldVer = outs[len(outs)-1]
				break
			}
		}

		newVer := ""
		if len(os.Args) > 2 {
			newVer = os.Args[2]
		} else {
			vers := strings.Split(oldVer, ".")
			v, err := strconv.Atoi(vers[len(vers)-1])
			if err != nil {
				v = 0
			}
			vers[len(vers)-1] = strconv.Itoa(v + 1)
			newVer = strings.Join(vers, ".")
		}

		fmt.Println("Upgrade", oldVer, "=>", newVer)
		outs, _ = runCommand("git", "tag", "-a", newVer, "-m", "by github.com/ssgo/tool/gomod")
		for _, line := range outs {
			fmt.Println(line)
		}
		fmt.Println("Pushing version tag ...")
		outs, _ = runCommand("git", "push", "origin", newVer)
		for _, line := range outs {
			fmt.Println(line)
		}
		fmt.Println("Done")

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	gomod")
	fmt.Println("	\033[36m-v\033[0m	\033[37m查看当前项目的版本列表\033[0m")
	fmt.Println("	\033[36m-u\033[0m	\033[37m版本号+1并提交\033[0m")
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	\033[36mgomod -v\033[0m")
	fmt.Println("	\033[36mgomod -u\033[0m")
	fmt.Println("	\033[36mgomod -u v1.2.1\033[0m")
	fmt.Println("")
}

func runCommand(command string, args ...string) ([]string, error) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	outs := make([]string, 0)
	cmd.Start()
	reader := bufio.NewReader(io.MultiReader(stdout, stderr))
	for {
		lineBuf, _, err2 := reader.ReadLine()

		if err2 != nil || io.EOF == err2 {
			break
		}
		line := strings.TrimRight(string(lineBuf), "\r\n")
		outs = append(outs, line)
	}

	cmd.Wait()
	return outs, nil
}
