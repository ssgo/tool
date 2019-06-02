package main

import (
	"bufio"
	"fmt"
	"github.com/ssgo/u"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var filesModTime = make(map[string]int64)

func main() {
	if len(os.Args) == 1 {
		printUsage()
		return
	}

	basePaths := make([]string, 0)
	cmd := "go"
	cmdArgs := make([]string, 0)
	var runArgs []string
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "help":
		case "--help":
			printUsage()
			return
		case "-p":
			i++
			tmpPaths := strings.Split(os.Args[i], ",")
			for _, path := range tmpPaths {
				if []byte(path)[len(path)-1] != '/' {
					path += "/"
				}
				basePaths = append(basePaths, path)
			}
		case "-sh":
			if i < len(os.Args)-1 && os.Args[i+1][0] != '-' {
				i++
				cmd = os.Args[i]
			} else {
				cmd = "sh"
			}
		case "-r":
			cmdArgs = append(cmdArgs, "run")
			runArgs = make([]string, 0)
			files, err := ioutil.ReadDir("./")
			if err == nil {
				for _, file := range files {
					if !strings.HasPrefix(file.Name(), ".") && !strings.HasSuffix(file.Name(), "_test.go") && strings.HasSuffix(file.Name(), ".go") {
						cmdArgs = append(cmdArgs, "./"+file.Name())
					}
				}
			}
		case "-t":
			fi, err := os.Stat("./tests")
			if err == nil && fi != nil {
				cmdArgs = append(cmdArgs, "test", "./tests")
			} else {
				cmdArgs = append(cmdArgs, "test", ".")
			}

		case "-b":
			cmdArgs = append(cmdArgs, "-bench", ".*")
		default:
			if runArgs == nil {
				cmdArgs = append(cmdArgs, os.Args[i])
			} else {
				runArgs = append(runArgs, os.Args[i])
			}
		}
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stop()
		fmt.Println("\nExit")
		os.Exit(0)
	}()

	if len(basePaths) == 0 {
		basePaths = append(basePaths, "./")
	}

	//os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
	//fmt.Printf("[Watching \033[36m%s\033[0m] [Running \033[36mgo %s\033[0m]\n\n", strings.Join(basePaths, " "), strings.Join(cmdArgs, " "))
	//runCommand("go", cmdArgs...)

	changed := make(chan bool)
	go func(changed chan bool) {
		for {
			if watchFiles() {
				stop()
				changed <- true
			}
			time.Sleep(time.Millisecond * 100)
		}
	}(changed)

	go func() {
		for {
			for _, path := range basePaths {
				watchPath(path)
			}
			time.Sleep(time.Second * 3)
		}
	}()

	for {
		select {
		case <-changed:
			_, _ = os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
			fmt.Printf("[Watching "+u.Cyan("%s")+" [Running "+u.Cyan("%s %s")+"\n\n", strings.Join(basePaths, " "), cmd, strings.Join(cmdArgs, " "))

			runPos := -1
			for i, arg := range cmdArgs {
				if arg == "run" {
					runPos = i
					break
				}
			}

			if runPos >= 0 && runArgs != nil {
				buildArgs := append([]string{}, cmdArgs[0:runPos]...)
				buildArgs = append(buildArgs, "build", "-o", ".run")
				buildArgs = append(buildArgs, cmdArgs[runPos+1:]...)
				fmt.Printf("Building "+u.Cyan("%s %s")+"\n", cmd, strings.Join(buildArgs, " "))
				runCommand(cmd, buildArgs...)
				runCommand("./.run", runArgs...)
			} else {
				runCommand(cmd, cmdArgs...)
			}
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	gowatch " + u.White("[-p paths] [-t] [-b] [...]"))
	fmt.Println("	" + u.Cyan("-p") + "	" + u.White("指定监视的路径，默认为 ./，支持逗号隔开的多个路径"))
	fmt.Println("	" + u.Cyan("-sh") + "	" + u.White("指定执行的命令，默认为 go"))
	fmt.Println("	" + u.Cyan("-r") + "	" + u.White("执行当前目录中的程序，相当于 go run *.go"))
	fmt.Println("	" + u.Cyan("-t") + "	" + u.White("执行测试用例，相当于 go test ./tests 或 go test ./tests（自动识别是否存在tests文件夹）"))
	fmt.Println("	" + u.Cyan("-b") + "	" + u.White("执行性能测试，相当于 go -bench .*，需要额外指定 -t 或 test 参数"))
	fmt.Println("	" + u.Cyan("...") + "	" + u.White("可以使用所有 go 命令的参数"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	" + u.Cyan("gowatch -r"))
	fmt.Println("	" + u.Cyan("gowatch -t"))
	fmt.Println("	" + u.Cyan("gowatch -t -b"))
	fmt.Println("	" + u.Cyan("gowatch -p ../ -t"))
	fmt.Println("	" + u.Cyan("gowatch run start.go"))
	fmt.Println("	" + u.Cyan("gowatch run samePackages start.go"))
	fmt.Println("	" + u.Cyan("gowatch test"))
	fmt.Println("	" + u.Cyan("gowatch test ./testcase"))
	fmt.Println("")
}

var lastCmd *exec.Cmd = nil

func runCommand(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), "GOGC=off")
	//cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	lastCmd = cmd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	_ = cmd.Start()
	reader := bufio.NewReader(io.MultiReader(stdout, stderr))
	//reader1 := bufio.NewReader(stdout)
	//reader2 := bufio.NewReader(stderr)
	for {
		lineBuf, _, err := reader.ReadLine()
		if err != nil || io.EOF == err {
			break
		}
		outputLine(string(lineBuf))

		//lineBuf1, _, err1 := reader1.ReadLine()
		//lineBuf2, _, err2 := reader2.ReadLine()
		//fmt.Println("##1", string(lineBuf1), err1)
		//fmt.Println("##2", string(lineBuf2), err2)
		//if err1 != nil || io.EOF == err1 || err2 != nil || io.EOF == err2 {
		//	break
		//}
		//outputLine(string(lineBuf1))
		//outputLine(string(lineBuf2))
	}

	_ = cmd.Wait()
	lastCmd = nil
}

func outputLine(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	if strings.HasPrefix(line, "ok ") {
		fmt.Println(u.BGreen(line))
	} else if strings.HasPrefix(line, "FAIL	") {
		fmt.Println(u.BRed(line))
	} else if strings.Index(line, ".go:") != -1 {
		fmt.Println(line)
		//if strings.Index(line, "go/src") != -1 {
		//	fmt.Println(line)
		//} else {
		//	fmt.Println(u.Cyan(line))
		//}
	} else if strings.HasPrefix(line, "	") {
		fmt.Println(line)
	} else {
		fmt.Println(line)
	}
}

func stop() {
	if lastCmd != nil {
		fmt.Println("killing ", lastCmd.Process.Pid)
		if runtime.GOOS == "windows" {
			_ = lastCmd.Process.Kill()
		} else {
			_ = lastCmd.Process.Signal(syscall.SIGTERM)
		}
		_, _ = lastCmd.Process.Wait()
		//syscall.Kill(-lastCmd.Process.Pid, syscall.SIGKILL)
	}
	_, err := os.Stat(".run")
	if err == nil || os.IsExist(err) {
		_ = os.Remove(".run")
	}
}

func watchFiles() bool {
	changed := false
	for fileName, modTime := range filesModTime {
		info, err := os.Stat(fileName)
		if err != nil {
			delete(filesModTime, fileName)
			continue
		}
		if info.ModTime().Unix() != modTime {
			filesModTime[fileName] = info.ModTime().Unix()
			changed = true
		}
	}
	return changed
}

func watchPath(path string) {
	allType := false
	if strings.HasSuffix(path, "*") {
		allType = true
		path = path[0 : len(path)-2]
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	for _, file := range files {
		fileName := file.Name()
		//fileBytes := []byte(file.Name())
		if fileName[0] == '.' {
			continue
		}
		if file.IsDir() {
			watchPath(path + file.Name() + "/")
		} else {
			if !allType && !strings.HasSuffix(fileName, ".go") && !strings.HasSuffix(fileName, ".json") && !strings.HasSuffix(fileName, ".yml") {
				continue
			}
			//l := len(fileBytes)
			//if l < 4 || fileBytes[l-3] != '.' || fileBytes[l-2] != 'g' || fileBytes[l-1] != 'o' {
			//	continue
			//}
			if filesModTime[path+file.Name()] == 0 {
				filesModTime[path+file.Name()] = 1
			}
		}
	}
}
