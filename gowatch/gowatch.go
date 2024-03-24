package main

import (
	"bufio"
	"fmt"
	"github.com/ssgo/u"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var filesModTimeLock = sync.Mutex{}
var filesModTime = make(map[string]int64)
var ignores = make([]string, 0)
var watchTypes = []string{".go", ".yml"}

func main() {
	if len(os.Args) == 1 {
		printUsage()
		return
	}

	if u.FileExists(".gitignore") {
		gitIgnores, _ := u.ReadFileLines(".gitignore")
		for _, line := range gitIgnores {
			if len(line) > 0 && line[0] == '/' {
				ignores = append(ignores, line[1:])
			}
		}
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
				//lastChar := []byte(path)[len(path)-1]
				//if lastChar != '/' && lastChar != '*' {
				//	path += "/"
				//}
				basePaths = append(basePaths, path)
			}
		case "-pt":
			i++
			tmpTypes := strings.Split(os.Args[i], ",")
			watchTypes = make([]string, 0)
			for _, typ := range tmpTypes {
				watchTypes = append(watchTypes, typ)
			}
		case "-ig":
			i++
			ignoreList := strings.Split(os.Args[i], ",")
			for _, igStr := range ignoreList {
				ignores = append(ignores, igStr)
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
			files, err := os.ReadDir("./")
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
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

	lastChanges := make([]string, 0)
	changedChannel := make(chan bool)
	go func(changedChannel chan bool) {
		for {
			if changed, changes := watchFiles(); changed {
				lastChanges = changes
				stop()
				changedChannel <- true
			}
			time.Sleep(time.Millisecond * 500)
		}
	}(changedChannel)

	go func() {
		for {
			for _, path := range basePaths {
				watchPath(path, false)
			}
			time.Sleep(time.Second * 3)
		}
	}()

	for {
		select {
		case <-changedChannel:
			_, _ = os.Stdout.WriteString("\x1b[3;J\x1b[H\x1b[2J")
			curPath, _ := os.Getwd()
			fmt.Printf("[at "+u.Dim("%s")+"]\n", curPath)
			fmt.Printf("[watching "+u.Cyan("%s")+" with "+u.Magenta("%s")+"]\n", strings.Join(basePaths, " "), strings.Join(watchTypes, " "))
			if len(ignores) > 0 {
				fmt.Printf("[ignores "+u.Yellow("%s")+"]\n", strings.Join(ignores, " "))
			}
			if len(lastChanges) > 0 {
				fmt.Println("[changed files]")
				for _, changedFile := range lastChanges {
					fmt.Println("  >", u.Yellow(changedFile))
				}
			}
			fmt.Printf("[running "+u.Cyan("%s %s")+"]\n\n", cmd, strings.Join(cmdArgs, " "))
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
	fmt.Println("	gowatch " + u.White("[-p paths] [-pt types] [-ig ignores] [-t] [-b] [...]"))
	fmt.Println("	" + u.Cyan("-p") + "	" + u.White("指定监视的路径，默认为 ./，支持逗号隔开的多个路径，以*结尾代表监听该文件夹下所有类型的文件"))
	fmt.Println("	" + u.Cyan("-pt") + "	" + u.White("指定监视的文件类型，默认为 .go,.yml 支持逗号隔开的多个类型"))
	fmt.Println("	" + u.Cyan("-ig") + "	" + u.White("排除指定的文件夹，默认从 .gitignore 中找到 / 开头的项目进行排除"))
	fmt.Println("	" + u.Cyan("-sh") + "	" + u.White("指定执行的命令，默认为 go"))
	fmt.Println("	" + u.Cyan("-r") + "	" + u.White("执行当前目录中的程序，相当于 go run *.go"))
	fmt.Println("	" + u.Cyan("-t") + "	" + u.White("执行测试用例，相当于 go test ./tests 或 go test ./tests（自动识别是否存在tests文件夹）"))
	fmt.Println("	" + u.Cyan("-b") + "	" + u.White("执行性能测试，相当于 go -bench .*，需要额外指定 -t 或 test 参数"))
	fmt.Println("	" + u.Cyan("...") + "	" + u.White("可以使用除 run 外的 go 命令的参数"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	" + u.Cyan("gowatch -r"))
	fmt.Println("	" + u.Cyan("gowatch -t"))
	fmt.Println("	" + u.Cyan("gowatch -t -b"))
	fmt.Println("	" + u.Cyan("gowatch -p ../ -t"))
	//fmt.Println("	" + u.Cyan("gowatch run start.go"))
	//fmt.Println("	" + u.Cyan("gowatch run samePackages start.go"))
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

func watchFiles() (bool, []string) {
	changed := false
	changes := make([]string, 0)
	filesModTimeLock.Lock()
	for filename, modTime := range filesModTime {
		info, err := os.Stat(filename)
		if err != nil {
			delete(filesModTime, filename)
			continue
		}
		if info.ModTime().Unix() != modTime {
			filesModTime[filename] = info.ModTime().Unix()
			if modTime > 1 {
				changes = append(changes, filename)
			}
			changed = true
		}
	}
	filesModTimeLock.Unlock()
	return changed, changes
}

func checkInType(filename string) bool {
	for _, typ := range watchTypes {
		if strings.HasSuffix(filename, typ) {
			return true
		}
	}
	return false
}

func watchPath(parent string, allType bool) {
	if strings.HasSuffix(parent, string(os.PathSeparator)+"*") {
		allType = true
		parent = parent[0 : len(parent)-2]
	}
	fileInfo, err := os.Stat(parent)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		files, err := os.ReadDir(parent)
		if err != nil {
			return
		}

		for _, file := range files {
			filename := file.Name()
			//fileBytes := []byte(file.Name())
			if filename[0] == '.' {
				continue
			}
			fullFilename := filepath.Join(parent, filename)
			ignored := false
			for _, ignore := range ignores {
				if strings.HasPrefix(fullFilename, ignore) {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}

			if file.IsDir() {
				//fmt.Println("watch path:", fullFilename, allType)
				watchPath(fullFilename, allType)
			} else {
				//fmt.Println("watch file:", filepath.Join(parent,filename), allType)
				if !allType && !checkInType(filename) {
					continue
				}
				//l := len(fileBytes)
				//if l < 4 || fileBytes[l-3] != '.' || fileBytes[l-2] != 'g' || fileBytes[l-1] != 'o' {
				//	continue
				//}
				fullFileName := filepath.Join(parent, file.Name())
				filesModTimeLock.Lock()
				if filesModTime[fullFileName] == 0 {
					filesModTime[fullFileName] = 1
				}
				filesModTimeLock.Unlock()
			}
		}
	} else {
		filesModTimeLock.Lock()
		if filesModTime[parent] == 0 {
			filesModTime[parent] = 1
		}
		filesModTimeLock.Unlock()
	}
}
