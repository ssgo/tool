package main

import (
	"bufio"
	"fmt"
	"github.com/ssgo/u"
	"io"
	"io/ioutil"
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
		outs, _ := runCommand("git", "tag", "-l", "v*", "--sort=taggerdate")
		for _, line := range outs {
			fmt.Println(line)
		}
		fmt.Println(len(outs), "Versions")
	case "-l", "v*", "--sort=taggerdate":
		path := "./"
		if len(os.Args) > 2 {
			path = os.Args[2]
			if len(path) > 0 && path[len(path)-1] != '/' {
				path += "/"
			}
		}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			fmt.Println(u.Red(err.Error()))
			return
		}
		os.Chdir(path)
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' {
				continue
			}

			lastVersion := ""
			os.Chdir(fileName)
			outs, _ := runCommand("git", "tag", "-l", "v*", "--sort=taggerdate")
			os.Chdir("..")
			for i := len(outs) - 1; i >= 0; i-- {
				if outs[i][0] == 'v' && strings.IndexByte(outs[i], '.') != -1 {
					lastVersion = outs[len(outs)-1]
					break
				}
			}
			if lastVersion != "" {
				fmt.Println(u.Cyan(fmt.Sprintf("%12s", fileName)), lastVersion)
			}
		}
	case "-u":
		oldVer := "v0.0.0"
		outs, _ := runCommand("git", "tag", "-l", "v*", "--sort=taggerdate")
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

	case "-c", "-cf", "-fc":
		force := len(os.Args[1]) > 2 || (len(os.Args) > 2 && os.Args[2] == "-f")
		goPathLines, _ := runCommand("go", "env", "GOPATH")
		goPath := "~/go/"
		if len(goPathLines) > 0 {
			goPath = goPathLines[0] + "/"
		}
		err := os.MkdirAll(goPath+"gomod/checks", 0755)
		if err != nil {
			fmt.Println(err)
			return
		}

		mods := map[string]string{}
		modLines, _ := readFile("go.mod")
		modReadState := 0
		for _, line := range modLines {
			line = strings.TrimSpace(line)
			if line == "require (" {
				modReadState = 1
				continue
			}
			if modReadState == 1 {
				if line == ")" {
					break
				}
				kv := strings.Split(line, " ")
				if len(kv) != 2 {
					continue
				}
				if strings.Index(kv[0], "golang.org") != -1 {
					continue
				}
				mods[kv[0]] = kv[1]
			}
		}

		lastMods := map[string]string{}
		maxModLen := 0
		for mod := range mods {
			_ = os.Chdir(goPath + "gomod/checks")

			if len(mod) > maxModLen {
				maxModLen = len(mod)
			}
			modPaths := strings.Split(mod, "/")
			modName := modPaths[len(modPaths)-1]

			if force {
				err = os.RemoveAll(modName + ".git")
				if err != nil {
					fmt.Println(err)
				}
			}

			if fileExists(modName + ".git") {
				_ = os.Chdir(modName + ".git")
			} else {
				fmt.Println(u.Cyan("cloning " + mod))
				err = printCommand("git", "clone", "--bare", "https://"+mod)
				if err != nil {
					fmt.Println(err)
				}
				_ = os.Chdir(modName + ".git")
			}

			lastVer := ""
			outs, _ := runCommand("git", "tag", "-l", "v*", "--sort=taggerdate")
			for i := len(outs) - 1; i >= 0; i-- {
				if outs[i][0] == 'v' && strings.IndexByte(outs[i], '.') != -1 {
					lastVer = outs[len(outs)-1]
					break
				}
			}

			lastMods[mod] = lastVer
		}

		for mod, ver := range mods {
			lastVer := lastMods[mod]
			if lastVer == ver {
				fmt.Printf(fmt.Sprint("%-", maxModLen+1, "s %s\n"), mod, u.BGreen(ver))
			} else {
				fmt.Printf(fmt.Sprint("%-", maxModLen+1, "s %s => %s\n"), mod, u.BRed(ver), u.Green(lastVer))
			}
		}

	default:
		if len(os.Args) > 1 {
			if os.Args[1] == "tidy" {
				_ = os.Remove("go.sum")
			}
			args := []string{"mod"}
			args = append(args, os.Args[1:]...)
			outs, _ := runCommand("go", args...)
			for _, line := range outs {
				fmt.Println(line)
			}
		} else {
			printUsage()
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	gomod")
	fmt.Println("	\033[36m-v\033[0m	\033[37m查看当前项目的版本列表\033[0m")
	fmt.Println("	\033[36m-l\033[0m	\033[37m查看当前子目录项目的最新版本\033[0m")
	fmt.Println("	\033[36m-u\033[0m	\033[37m版本号+1并提交\033[0m")
	fmt.Println("	\033[36m-c [-f]\033[0m	\033[37m检查 go.mod 中依赖的包的最新版本，-f会强制更新已缓存版本\033[0m")
	fmt.Println("	\033[36minit tidy download vendor verify why help\033[0m	\033[37m等同于 go mod ...\033[0m")
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	\033[36mgomod -v\033[0m")
	fmt.Println("	\033[36mgomod -l\033[0m")
	fmt.Println("	\033[36mgomod -u\033[0m")
	fmt.Println("	\033[36mgomod -u v1.2.1\033[0m")
	fmt.Println("	\033[36mgomod -c\033[0m")
	fmt.Println("	\033[36mgomod -c -f\033[0m")
	fmt.Println("	\033[36mgomod init ...\033[0m")
	fmt.Println("	\033[36mgomod tidy\033[0m")
	fmt.Println("	\033[36mgomod download\033[0m")
	fmt.Println("	\033[36mgomod vendor\033[0m")
	fmt.Println("	\033[36mgomod verify\033[0m")
	fmt.Println("	\033[36mgomod why\033[0m")
	fmt.Println("	\033[36mgomod help\033[0m")
	fmt.Println("")
}

func printCommand(command string, args ...string) error {
	fmt.Println(command, strings.Join(args, " "))
	lines, err := runCommand(command, args...)
	for _, line := range lines {
		fmt.Println(line)
	}
	return err
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

func readFile(fileName string) ([]string, error) {
	outs := make([]string, 0)
	fd, err := os.OpenFile("go.mod", os.O_RDONLY, 0400)
	if err != nil {
		return outs, err
	}

	inputReader := bufio.NewReader(fd)
	for {
		line, err := inputReader.ReadString('\n')
		line = strings.TrimRight(string(line), "\r\n")
		outs = append(outs, line)
		if err != nil {
			break
		}
	}
	fd.Close()
	return outs, nil
}

func fileExists(fileName string) bool {
	fi, err := os.Stat(fileName)
	return err == nil && fi != nil
}
