package main

import (
	"encoding/json"
	"fmt"
	"github.com/ssgo/log"
	"github.com/ssgo/standard"
	"github.com/ssgo/u"
	"io"
	"math"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// var useJson bool
var showShortTime bool

func main() {
	fileName := ""
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "help":
		case "--help":
			printUsage()
			return
		//case "-j":
		//	useJson = true
		case "-s":
			showShortTime = true
		default:
			if fileName == "" {
				fileName = os.Args[i]
			}
		}
	}

	var fd *os.File = nil

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		//if fd != nil {
		//	_ = fd.Close()
		//}
	}()

	if fileName != "" {
		var err error
		fd, err = os.Open(fileName)
		if err != nil {
			fmt.Println(u.Red(err.Error()))
			return
		}
	} else {
		fd = os.Stdin
	}

	//inputReader := bufio.NewReader(fd)
	inputChan := make(chan string)
	go func() {
		buf := make([]byte, 4096)
		for {
			n1, err := fd.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				inputChan <- "__EOF__"
				break
			}
			inputChan <- string(buf[0:n1])
		}
	}()

	savedLine := ""
	for {
		line := ""
		withWrap := true
		if savedLine != "" {
			// 之前有数据
			pos := strings.IndexByte(savedLine, '\n')
			if pos != -1 {
				// 之前有未处理的行
				line = savedLine[0:pos]
				savedLine = savedLine[pos+1:]
			}
		}
		if line == "" {
			// 读取
			str := ""
			select {
			case str = <-inputChan:
				//fmt.Println("<<", u.BYellow(str))
			case <-time.After(1 * time.Microsecond * 100):
				withWrap = false
			}
			if str == "__EOF__" {
				break
			}

			if str != "" {
				// 读到数据
				pos := strings.IndexByte(str, '\n')
				if pos != -1 {
					// 读到 \n
					line = savedLine + str[0:pos]
					savedLine = str[pos+1:]
				} else {
					// 未读到 \n
					savedLine += str
					continue
				}
			} else {
				// 超时先输出
				line = savedLine
				savedLine = ""
			}
		}

		line = strings.TrimRight(line, "\r\n")
		output(line, withWrap)
	}
}

func shortTime(tm string) string {
	return strings.Replace(tm[5:16], "T", " ", 1)
}

type LevelOutput struct {
	level    string
	levelKey string
}

func (levelOutput *LevelOutput) Print(v string) {
	switch strings.ToLower(levelOutput.level) {
	case "debug":
		fmt.Print(" ", v)
	case "info":
		fmt.Print(" ", u.Cyan(v))
	case "warning":
		fmt.Print(" ", u.Yellow(v))
	case "error":
		fmt.Print(" ", u.Red(v))
	}
	return
}

func (levelOutput *LevelOutput) BPrint(v string) {
	switch strings.ToLower(levelOutput.level) {
	case "debug":
		fmt.Print(u.BWhite(v))
	case "info":
		fmt.Print(u.BCyan(v))
	case "warning":
		fmt.Print(u.BYellow(v))
	case "error":
		fmt.Print(u.BRed(v))
	}
	return
}

var errorLineMatcher = regexp.MustCompile("(\\w+\\.go:\\d+)")
var lastOutputIsWrap = false

func output(line string, withWrap bool) {
	if line == "" {
		if !lastOutputIsWrap {
			fmt.Println()
		}
		lastOutputIsWrap = true
		return
	}
	lastOutputIsWrap = withWrap

	b := log.ParseBaseLog(line)
	//fmt.Println(u.JsonP(b))
	if b == nil {
		// 高亮错误代码
		if strings.Contains(line, ".go:") {
			if strings.Contains(line, "/ssgo/") || strings.Contains(line, "/ssdo/") {
				line = errorLineMatcher.ReplaceAllString(line, u.BYellow("$1"))
			} else if !strings.Contains(line, "/go/src/") {
				line = errorLineMatcher.ReplaceAllString(line, u.BRed("$1"))
			}
		}
		if withWrap {
			fmt.Println(line)
		} else {
			fmt.Print(line)
		}
		return
	}

	var logTime time.Time
	if strings.ContainsRune(b.LogTime, 'T') {
		logTime = log.MakeTime(b.LogTime)
	} else {
		ft := u.Float64(b.LogTime)
		ts := int64(math.Floor(ft))
		tns := int64((ft - float64(ts)) * 1e9)
		logTime = time.Unix(ts, tns)
	}

	showTime := logTime.Format(u.StringIf(!showShortTime, "2006-01-02 15:04:05.000000", "01-02 15:04:05"))

	t1 := strings.Split(showTime, " ")
	d := t1[0]
	t := ""
	if len(t1) > 1 {
		t = t1[1]
	}
	t2 := strings.Split(t, ".")
	s := ""
	if len(t2) > 1 {
		s = t2[1]
	}
	t = t2[0]
	fmt.Print(u.BWhite(d + " " + t))
	if s != "" {
		fmt.Print(u.White("." + s))
	}
	fmt.Print(" ", u.White(b.TraceId, u.AttrDim, u.AttrUnderline))

	lo := LevelOutput{}
	if b.Extra["debug"] != nil {
		lo.level = "debug"
		lo.levelKey = "debug"
	} else if b.Extra["warning"] != nil {
		lo.level = "warning"
		lo.levelKey = "warning"
	} else if b.Extra["error"] != nil {
		lo.level = "error"
		lo.levelKey = "error"
	} else if b.Extra["info"] != nil {
		lo.level = "info"
		lo.levelKey = "info"
	} else if b.Extra["Debug"] != nil {
		lo.level = "debug"
		lo.levelKey = "Debug"
	} else if b.Extra["Warning"] != nil {
		lo.level = "warning"
		lo.levelKey = "Warning"
	} else if b.Extra["Error"] != nil {
		lo.level = "error"
		lo.levelKey = "Error"
	} else if b.Extra["Info"] != nil {
		lo.level = "info"
		lo.levelKey = "Info"
	}

	if b.LogType == standard.LogTypeRequest {
		r := standard.RequestLog{}
		log.ParseSpecialLog(b, &r)
		if r.ResponseCode <= 0 || (r.ResponseCode >= 400 && r.ResponseCode <= 599) {
			fmt.Print(" ", u.BRed(u.String(r.ResponseCode)), " ", u.Red(u.String(r.UsedTime)))
		} else {
			fmt.Print(" ", u.BGreen(u.String(r.ResponseCode)), " ", u.Green(u.String(r.UsedTime)))
		}

		fmt.Print("  ", r.ClientIp, u.Dim(" from"), u.Dim("("), u.Cyan(r.FromApp), u.Dim(":"), r.FromNode, u.Dim(")"), u.Dim(" to"), u.Dim("("), u.Cyan(r.App), u.Dim(":"), r.Node, u.Dim(":"), r.AuthLevel, u.Dim(":"), r.Priority, u.Dim(")"))
		if r.RequestId != r.TraceId {
			fmt.Print(u.Dim("  requestId:"), r.RequestId)
		}
		fmt.Print("  ", u.Dim("user"), u.Dim(":"), u.Cyan(r.UserId), u.Dim(" sess"), u.Dim(":"), u.Cyan(r.SessionId), u.Dim(" dev"), u.Dim(":"), u.Cyan(r.DeviceId), u.Dim(" app"), u.Dim(":"), u.Cyan(r.ClientAppName), u.Dim(":"), u.Cyan(r.ClientAppVersion))
		fmt.Print("  ", r.Scheme, " ", r.Proto, " ", r.Host, " ", r.Method, " ", u.Cyan(r.Path))
		if r.RequestData != nil {
			for k, v := range r.RequestData {
				fmt.Print("  ", u.Cyan(k, u.AttrItalic), u.Dim(":"), u.String(v))
			}
		}
		if !showShortTime {
			if r.RequestHeaders != nil {
				for k, v := range r.RequestHeaders {
					fmt.Print("  ", u.Cyan(k, u.AttrDim, u.AttrItalic), u.Dim(":"), u.String(v))
				}
			}
		}

		fmt.Print("  ", u.BWhite(u.String(r.ResponseDataLength)))
		if !showShortTime {
			if r.ResponseHeaders != nil {
				for k, v := range r.ResponseHeaders {
					fmt.Print("  ", u.Blue(k, u.AttrDim, u.AttrItalic), u.Dim(":"), u.String(v))
				}
			}
			fmt.Print("  ", u.String(r.ResponseData))
		}
	} else if b.LogType == standard.LogTypeStatistic {
		r := standard.StatisticLog{}
		log.ParseSpecialLog(b, &r)
		fmt.Print(" ", u.Cyan(r.Name, u.AttrBold))
		fmt.Print("  ", u.Dim(r.App))
		fmt.Print(" ", u.Dim(shortTime(r.StartTime)+" ~ "+shortTime(r.EndTime)))
		fmt.Print(" ", u.Green(u.String(r.Times)), " ", u.Magenta(u.String(r.Failed)))
		fmt.Print(" ", fmt.Sprintf("%.4f", r.Min), " ", u.Cyan(fmt.Sprintf("%.4f", r.Avg)), " ", fmt.Sprintf("%.4f", r.Max))
	} else if b.LogType == standard.LogTypeTask {
		r := standard.TaskLog{}
		log.ParseSpecialLog(b, &r)
		if r.Succeed {
			fmt.Print("  ", u.Green(r.Name), " ", u.BGreen(fmt.Sprintf("%.4f", r.UsedTime)))
		} else {
			fmt.Print("  ", u.Red(r.Name), " ", u.BRed(fmt.Sprintf("%.4f", r.UsedTime)))
		}
		fmt.Print(" @", u.Dim(shortTime(r.StartTime)), " @", u.Dim(r.Node))
		fmt.Print(" ", u.Json(r.Args))
		fmt.Print(" ", u.Magenta(r.Memo))
	} else {
		if lo.level != "" {
			lo.Print(u.String(b.Extra[lo.levelKey]))
			delete(b.Extra, lo.levelKey)
		} else if b.LogType == "undefined" {
			fmt.Print(" ", u.Dim("-"))
		} else {
			fmt.Print(" ", u.Cyan(b.LogType, u.AttrBold))
		}
	}

	callStacks := b.Extra["callStacks"]
	if callStacks != nil {
		delete(b.Extra, "callStacks")
	}

	var codeFileMatcher = regexp.MustCompile(`(\w+?\.)(go|js)`)
	if b.Extra != nil {
		for k, v := range b.Extra {
			if k == "extra" && u.String(v)[0] == '{' {
				extra := map[string]interface{}{}
				u.UnJson(u.String(v), &extra)
				for k2, v2 := range extra {
					v2Str := u.String(v2)
					if k2 == "stack" && v2Str != "" {
						fmt.Println()
						for _, line := range u.SplitWithoutNone(v2Str, "\n") {
							a := strings.Split(line, "```")
							for i := 0; i < len(a); i++ {
								if i%2 == 0 {
									a[i] = codeFileMatcher.ReplaceAllString(a[i], u.BRed("$1$2"))
								} else {
									a[i] = u.BCyan(a[i])
								}
							}
							fmt.Println("  ", strings.Join(a, ""))
						}
					} else {
						fmt.Print("  ", u.White(k2+":", u.AttrDim, u.AttrItalic), v2Str)
					}
				}
			} else {
				fmt.Print("  ", u.White(k+":", u.AttrDim, u.AttrItalic), u.String(v))
			}
		}
	}

	if !showShortTime && callStacks != nil {
		var callStacksList []interface{}
		if callStacksStr, ok := callStacks.(string); ok && len(callStacksStr) > 2 && callStacksStr[0] == '[' {
			callStacksList = make([]interface{}, 0)
			json.Unmarshal([]byte(callStacksStr), &callStacksList)
		} else {
			callStacksList, ok = callStacks.([]interface{})
		}

		if callStacksList != nil && len(callStacksList) > 0 {
			fmt.Println()
			for _, vi := range callStacksList {
				v := u.String(vi)
				postfix := ""
				if pos := strings.LastIndexByte(v, '/'); pos != -1 {
					postfix = v[pos+1:]
					v = v[0 : pos+1]
				} else {
					postfix = v
					v = ""
				}
				fmt.Print(" ", u.Dim(v))
				if len(v) > 2 && (v[0] == '/' || v[1] == ':') {
					lo.BPrint(postfix)
				} else {
					lo.Print(postfix)
				}
				fmt.Println()
			}
		} else {
			fmt.Print(" ")
			lo.Print(u.String(callStacks))
		}
	}
	if withWrap {
		fmt.Println()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	logv [-j] [-s] [file]")
	fmt.Println("	" + u.Cyan("-j") + "	" + u.White("josn output"))
	fmt.Println("	" + u.Cyan("-s") + "	" + u.White("show full time"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	" + u.Cyan("tail ***.log | logv"))
	fmt.Println("	" + u.Cyan("logv ***.log"))
	fmt.Println("	" + u.Cyan("tail ***.log | logv -j -f"))
	fmt.Println("")
}
