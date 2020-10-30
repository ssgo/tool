package main

import (
	"bufio"
	"fmt"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
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
	case "-l":
		rootPath := "./"
		if len(os.Args) > 2 {
			rootPath = os.Args[2]
			if len(rootPath) > 0 && rootPath[len(rootPath)-1] != '/' {
				rootPath += "/"
			}
		}
		files, err := ioutil.ReadDir(rootPath)
		if err != nil {
			fmt.Println(u.Red(err.Error()))
			return
		}
		os.Chdir(rootPath)
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' {
				continue
			}

			fi, _ := os.Stat(fileName)
			if !fi.IsDir() {
				continue
			}

			if !u.FileExists(path.Join(fileName, ".git")){
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
		commit := ""
		if len(os.Args) > 3 {
			commit = os.Args[3]
		}

		if commit != "" {
			_ = printCommand("git", "commit", "-a", "-m", commit)
			_ = printCommand("git", "push")
		}
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

	case "-c", "-cf":
		cachePath := fmt.Sprintf("%s%cgomodCache%c", os.TempDir(), os.PathSeparator, os.PathSeparator)
		_ = os.Mkdir(cachePath, 0755)

		force := len(os.Args[1]) > 2 || (len(os.Args) > 2 && os.Args[2] == "-f")
		if force {
			fmt.Println("cache path: ", cachePath)
			_ = os.RemoveAll(cachePath)
		}

		mods := map[string]string{}
		modLines, _ := readFile("go.mod")
		modReadState := 0
		for _, line := range modLines {
			line = strings.TrimSpace(line)
			if line == "require (" {
				modReadState = 1
				continue
			} else if strings.HasPrefix(line, "require ") {
				modReadState = 1
				line = line[8:]
			}
			if modReadState == 1 {
				if line == ")" {
					break
				}
				kv := strings.Split(line, " ")
				if len(kv) != 2 {
					continue
				}
				if strings.Index(kv[0], "github.com/") == -1 {
					continue
				}
				mods[strings.Replace(kv[0], "github.com/", "", 1)] = kv[1]
			}
		}

		lastMods := map[string]string{}
		maxModLen := 0
		hc := httpclient.GetClient(10 * time.Second)
		for mod := range mods {
			cacheFile := cachePath + strings.Replace(mod, "/", "_", 20)
			fs, err := os.Stat(cacheFile)
			isOk := false
			if fs != nil && err == nil && fs.ModTime().Unix() >= time.Now().Add(-300 * time.Second).Unix() {
				ver := ""
				err := u.Load(cacheFile, &ver)
				if err != nil {
					log.DefaultLogger.Error(err.Error())
				} else {
					lastMods[mod] = ver
					isOk = true
				}
			}

			if !isOk {
				r := make([]struct{ Name string }, 0)
				fmt.Println("fetching: https://api.github.com/repos/" + mod + "/tags?per_page=1")
				err := hc.Get("https://api.github.com/repos/" + mod + "/tags?per_page=1").To(&r)
				if err != nil {
					log.DefaultLogger.Error(err.Error())
					return
				}
				if err == nil && len(r) > 0 && r[0].Name != "" {
					lastMods[mod] = r[0].Name
					err = u.Save(cacheFile, r[0].Name)
					if err != nil {
						log.DefaultLogger.Error(err.Error())
					}
				}
			}
		}

		//force := len(os.Args[1]) > 2 || (len(os.Args) > 2 && os.Args[2] == "-f")
		//goPathLines, _ := runCommand("go", "env", "GOPATH")
		//goPath := "~/go/"
		//if len(goPathLines) > 0 {
		//	goPath = goPathLines[0] + "/"
		//}
		//err := os.MkdirAll(goPath+"gomod/checks", 0755)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//
		//mods := map[string]string{}
		//modLines, _ := readFile("go.mod")
		//modReadState := 0
		//for _, line := range modLines {
		//	line = strings.TrimSpace(line)
		//	if line == "require (" {
		//		modReadState = 1
		//		continue
		//	}
		//	if modReadState == 1 {
		//		if line == ")" {
		//			break
		//		}
		//		kv := strings.Split(line, " ")
		//		if len(kv) != 2 {
		//			continue
		//		}
		//		if strings.Index(kv[0], "golang.org") != -1 {
		//			continue
		//		}
		//		mods[kv[0]] = kv[1]
		//	}
		//}
		//
		//lastMods := map[string]string{}
		//maxModLen := 0
		//for mod := range mods {
		//	_ = os.Chdir(goPath + "gomod/checks")
		//
		//	if len(mod) > maxModLen {
		//		maxModLen = len(mod)
		//	}
		//	modPaths := strings.Split(mod, "/")
		//	modName := modPaths[len(modPaths)-1]
		//
		//	if force {
		//		err = os.RemoveAll(modName + ".git")
		//		if err != nil {
		//			fmt.Println(err)
		//		}
		//	}
		//
		//	if fileExists(modName + ".git") {
		//		_ = os.Chdir(modName + ".git")
		//	} else {
		//		fmt.Println(u.Cyan("cloning " + mod))
		//		err = printCommand("git", "clone", "--bare", "https://"+mod)
		//		if err != nil {
		//			fmt.Println(err)
		//		}
		//		_ = os.Chdir(modName + ".git")
		//	}
		//
		//	lastVer := ""
		//	outs, _ := runCommand("git", "tag", "-l", "v*", "--sort=taggerdate")
		//	for i := len(outs) - 1; i >= 0; i-- {
		//		if outs[i][0] == 'v' && strings.IndexByte(outs[i], '.') != -1 {
		//			lastVer = outs[len(outs)-1]
		//			break
		//		}
		//	}
		//
		//	lastMods[mod] = lastVer
		//}

		for mod, ver := range mods {
			lastVer := lastMods[mod]
			if lastVer == ver {
				fmt.Printf(fmt.Sprint("%-", maxModLen+1, "s %s => %s\n"), mod, u.BGreen(ver), u.Green(lastVer))
			} else {
				fmt.Printf(fmt.Sprint("%-", maxModLen+1, "s %s => %s\n"), mod, u.BRed(ver), u.Green(lastVer))
			}
		}

	default:
		if len(os.Args) > 1 {
			args := []string{"mod"}
			args = append(args, os.Args[1:]...)
			outs, _ := runCommand("go", args...)
			for _, line := range outs {
				fmt.Println(line)
			}
			if os.Args[1] == "tidy" {
				_ = os.Remove("go.sum")
			}
		} else {
			printUsage()
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	gomod")
	fmt.Println("	" + u.Cyan("-v") + "	" + u.White("查看当前项目的版本列表"))
	fmt.Println("	" + u.Cyan("-l") + "	" + u.White("查看当前子目录项目的最新版本"))
	fmt.Println("	" + u.Cyan("-u") + "	" + u.White("版本号+1并提交"))
	fmt.Println("	" + u.Cyan("-c  [-f]") + "	" + u.White("检查 go.mod 中依赖的包的最新版本，-f会强制更新已缓存版本"))
	fmt.Println("	" + u.Cyan("init tidy download vendor verify why help") + "	" + u.White("等同于 go mod ..."))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	" + u.Cyan("gomod -v"))
	fmt.Println("	" + u.Cyan("gomod -l"))
	fmt.Println("	" + u.Cyan("gomod -u"))
	fmt.Println("	" + u.Cyan("gomod -u v1.2.1"))
	fmt.Println("	" + u.Cyan("gomod -c"))
	fmt.Println("	" + u.Cyan("gomod -c -f"))
	fmt.Println("	" + u.Cyan("gomod init ..."))
	fmt.Println("	" + u.Cyan("gomod tidy"))
	fmt.Println("	" + u.Cyan("gomod download"))
	fmt.Println("	" + u.Cyan("gomod vendor"))
	fmt.Println("	" + u.Cyan("gomod verify"))
	fmt.Println("	" + u.Cyan("gomod why"))
	fmt.Println("	" + u.Cyan("gomod help"))
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
