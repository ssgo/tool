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

var useJson bool
var showFullTime bool

func main() {
	fileName := ""
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "help":
		case "--help":
			printUsage()
			return
		case "-j":
			useJson = true
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
			fd.Close()
		}
		//os.Exit(0)
	}()

	if fileName != "" {
		var err error
		fd, err = os.Open(fileName)
		if err != nil {
			fmt.Println("\033[31m", err, "\033[0m")
			return
		}
	} else {
		fd = os.Stdin
	}

	inputReader := bufio.NewReader(fd)
	for {
		line, err := inputReader.ReadString('\n')
		//fmt.Println("[",line,"]")
		line = strings.TrimSpace(line)
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
	showTime := b.LogTime.Format(u.StringIf(showFullTime, "2006-01-02 15:04:05.000000", "01-02 15:04:05"))
	switch b.LogLevel {
	case "debug":
		showTime = u.Dim(showTime)
	case "info":
		showTime = showTime
	case "warning":
		showTime = u.BYellow(showTime)
	case "error":
		showTime = u.BRed(showTime)
	}
	fmt.Print(showTime)

	if b.LogType == standard.LogTypeRequest {
		r := log.RequestLog{}
		if log.ParseSpecialLog(b, &r) == nil {
			if r.ResponseCode <= 0 || (r.ResponseCode >= 400 && r.ResponseCode <= 599) {
				fmt.Print(" ", u.BRed(u.String(r.ResponseCode)), " ", u.Red(u.String(r.UsedTime)))
			} else {
				fmt.Print(" ", u.BGreen(u.String(r.ResponseCode)), " ", u.Green(u.String(r.UsedTime)))
			}

			fmt.Print("  ", u.Cyan(r.ClientIp), u.Dim("("), r.FromApp, u.Dim(":"), r.FromNode, u.Dim(")"), u.Dim(" => "), u.Cyan(r.App), u.Dim(":"), r.AuthLevel, u.Dim(":"), r.Priority, u.Dim("@"), r.Node)
			fmt.Print("  ", r.RequestId, u.Dim(":"), r.ClientId, u.Dim(":"), r.SessionId)
			fmt.Print("  ", r.Method, " ", r.Host, u.Cyan(r.Path))
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
		fmt.Print(" ", u.White(u.StringIf(b.LogType == "undefined", "-", b.LogType), u.AttrBold))

		levelMessage := b.Extra[b.LogLevel]
		if levelMessage != nil {
			delete(b.Extra, b.LogLevel)
			switch b.LogLevel {
			case "debug":
				fmt.Print("  ", u.Dim(levelMessage))
			case "info":
				fmt.Print("  ", levelMessage)
			case "warning":
				fmt.Print("  ", u.Yellow(levelMessage))
			case "error":
				fmt.Print("  ", u.Red(levelMessage))
			}
		}
	}

	if b.Extra != nil {
		for k, v := range b.Extra {
			fmt.Print("  ", u.White(k+":", u.AttrDim, u.AttrItalic), u.String(v))
		}
	}

	if b.Traces != "" {
		traces := strings.Split(b.Traces, "; ")
		fmt.Print(" ")
		for _, v := range traces {
			switch b.LogLevel {
			case "error":
				fmt.Print(" ", u.Red(v, u.AttrDim, u.AttrItalic))
			case "warning":
				fmt.Print(" ", u.Yellow(v, u.AttrDim, u.AttrItalic))
			default:
				fmt.Print(" ", u.White(v, u.AttrDim, u.AttrItalic))
			}
		}
	}
	fmt.Println()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	lv [-j] [-f] [file]")
	fmt.Println("	\033[36m-j\033[0m		\033[37mJosn output\033[0m")
	fmt.Println("	\033[36m-f\033[0m		\033[37mshow full time\033[0m")
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	\033[36mtail ***.log | lv\033[0m")
	fmt.Println("	\033[36mlv ***.log\033[0m")
	fmt.Println("	\033[36mtail ***.log | lv -j -f\033[0m")
	fmt.Println("")
}
