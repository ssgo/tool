package main

import (
	"bufio"
	"fmt"
	"github.com/ssgo/log"
	"github.com/ssgo/standard"
	"github.com/ssgo/u"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

//var useJson bool
var showFullTime bool

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
		case "-f":
			showFullTime = true
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
		if fd != nil {
			_ = fd.Close()
		}
		//os.Exit(0)
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

	inputReader := bufio.NewReader(fd)
	for {
		line, err := inputReader.ReadString('\n')
		//fmt.Println("[",line,"]")
		line = strings.TrimRight(line, "\r\n")
		output(line)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
	}
}

func output(line string) {
	if line == "" {
		return
	}

	b := log.ParseBaseLog(line)
	if b == nil {
		fmt.Println(line)
		return
	}
	logTime := log.MakeTime(b.LogTime)
	showTime := logTime.Format(u.StringIf(showFullTime, "2006-01-02 15:04:05.000000", "01-02 15:04:05"))
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
	fmt.Print(u.Dim(d), " ")
	if b.Extra["debug"] != nil {
		fmt.Print(u.Dim(t))
	} else if b.Extra["warning"] != nil {
		fmt.Print(u.BYellow(t))
	} else if b.Extra["error"] != nil {
		fmt.Print(u.BRed(t))
	}else{
		fmt.Print(t)
	}
	if s != "" {
		fmt.Print(u.Dim("."+s))
	}
	fmt.Print(" ", u.White(b.TraceId, u.AttrDim, u.AttrUnderline))

	if b.LogType == standard.LogTypeRequest {
		r := standard.RequestLog{}
		if log.ParseSpecialLog(b, &r) == nil {
			if r.ResponseCode <= 0 || (r.ResponseCode >= 400 && r.ResponseCode <= 599) {
				fmt.Print(" ", u.BRed(u.String(r.ResponseCode)), " ", u.Red(u.String(r.UsedTime)))
			} else {
				fmt.Print(" ", u.BGreen(u.String(r.ResponseCode)), " ", u.Green(u.String(r.UsedTime)))
			}

			fmt.Print("  ", u.Cyan(r.ClientIp), u.Dim("("), r.FromApp, u.Dim(":"), r.FromNode, u.Dim(")"), u.Dim(" => "), u.Cyan(r.App), u.Dim(":"), r.AuthLevel, u.Dim(":"), r.Priority, u.Dim("@"), r.Node)
			fmt.Print("  ", r.RequestId, u.Dim(":"), r.ClientId, u.Dim(":"), r.SessionId)
			fmt.Print("  ", r.Host, " ", r.Scheme, " ", r.Proto, " ", r.Method, " ", u.Cyan(r.Path))
			if r.RequestData != nil {
				for k, v := range r.RequestData {
					fmt.Print("  ", u.Magenta(k, u.AttrItalic), u.Dim(":"), u.String(v))
				}
			}
			if showFullTime {
				if r.RequestHeaders != nil {
					for k, v := range r.RequestHeaders {
						fmt.Print("  ", u.Magenta(k, u.AttrDim, u.AttrItalic), u.Dim(":"), u.String(v))
					}
				}
			}

			fmt.Print("  ", u.BWhite(u.String(r.ResponseDataLength)))
			if showFullTime {
				if r.ResponseHeaders != nil {
					for k, v := range r.ResponseHeaders {
						fmt.Print("  ", u.Blue(k, u.AttrDim, u.AttrItalic), u.Dim(":"), u.String(v))
					}
				}
				fmt.Print("  ", u.String(r.ResponseData))
			}
		}
	} else {
		if b.Extra["debug"] != nil {
			fmt.Print("  ", u.Dim(b.Extra["debug"]))
			delete(b.Extra, "debug")
		} else if b.Extra["info"] != nil {
			fmt.Print("  ", b.Extra["info"])
			delete(b.Extra, "info")
		} else if b.Extra["warning"] != nil {
			fmt.Print("  ", u.Yellow(b.Extra["warning"]))
			delete(b.Extra, "warning")
		} else if b.Extra["error"] != nil {
			fmt.Print("  ", u.Red(b.Extra["error"]))
			delete(b.Extra, "error")
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

	if b.Extra != nil {
		for k, v := range b.Extra {
			fmt.Print("  ", u.White(k+":", u.AttrDim, u.AttrItalic), u.String(v))
		}
	}

	if showFullTime && callStacks != nil {
		callStacksList, ok := callStacks.([]interface{})

		fmt.Print(" ")
		if ok {
			for _, v := range callStacksList {
				fmt.Print(" ", u.Magenta(v, u.AttrItalic))
			}
		} else {
			fmt.Print(" ", u.Magenta(u.String(callStacks), u.AttrItalic))
		}
	}
	fmt.Println()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	lv [-j] [-f] [file]")
	fmt.Println("	" + u.Cyan("-j") + u.White("Josn output"))
	fmt.Println("	" + u.Cyan("-f") + u.White("show full time"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	" + u.Cyan("tail ***.log | lv"))
	fmt.Println("	" + u.Cyan("lv ***.log"))
	fmt.Println("	" + u.Cyan("tail ***.log | lv -j -f"))
	fmt.Println("")
}
