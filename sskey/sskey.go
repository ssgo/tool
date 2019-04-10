package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/ssgo/u"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		printUsage()
		return
	}

	homeDir, _ := os.UserHomeDir()
	keyPath := fmt.Sprintf("%s%csskeys%c", homeDir, os.PathSeparator, os.PathSeparator)
	os.Mkdir(keyPath, 0700)

	op := os.Args[1]
	if (op == "-c" || op == "-t" || op == "-e" || op == "-d" || op == "-o") && len(os.Args) < 3 {
		keyName := scanLine("\033[36mPlease enter key name:\033[0m ")
		if keyName == "" {
			printUsage()
			fmt.Println("\033[31mneed key\033[0m")
			return
		}
		os.Args = append(os.Args, keyName)
	}

	if (op == "-e" || op == "-d") && len(os.Args) < 4 {
		data := scanLine("\033[36mPlease enter data:\033[0m ")
		if data == "" {
			printUsage()
			fmt.Println("\033[31mneed data\033[0m")
			return
		}
		os.Args = append(os.Args, data)
	}

	switch op {
	case "-l":
		files, err := ioutil.ReadDir(keyPath)
		if err != nil {
			fmt.Println("\033[31", err, "\033[0m")
			return
		}
		n := 0
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' {
				continue
			}
			n++
			fmt.Print("	\033[36m", fileName, "\033[0m	\033[37m", keyPath, fileName, "\033[0m\n")
		}
		fmt.Println(n, "Keys")
	case "-c":
		keyName := os.Args[2]

		fi, err := os.Stat(keyPath + keyName)
		if err == nil && fi != nil {
			fmt.Println("\033[31mkey exists\033[0m")
			return
		}

		fd, err := os.OpenFile(keyPath+keyName, os.O_CREATE|os.O_WRONLY, 0400)
		if err != nil {
			fmt.Println("\033[31mbad key file\033[0m")
			fmt.Println("\033[31m", err, "\033[0m")
			return
		}

		buf := make([]byte, 81)
		for i := 0; i < 40; i++ {
			buf[i] = byte(u.GlobalRand1.Intn(255))
			buf[40+i] = byte(u.GlobalRand2.Intn(255))
		}
		buf[80] = 217
		fd.WriteString(base64.StdEncoding.EncodeToString(buf))
		fd.Close()

		key, iv := loadKey(keyPath + os.Args[2])
		s1 := u.EncryptAes("Hello World!", key[2:], iv[5:])
		s2 := u.DecryptAes(s1, key[2:], iv[5:])
		fmt.Println("\033[36m", keyName, "\033[0m Created at", keyPath+keyName)
		fmt.Println("  Test Encrypt: \033[33mHello World!\033[0m => \033[33m", s1, "\033[0m")
		fmt.Println("  Test Decrypt: \033[33m", s1, "\033[0m => \033[33m", s2, "\033[0m")

	case "-t":
		key, iv := loadKey(keyPath + os.Args[2])
		s := "你好，Hello，안녕하세요，こんにちは，ON LI DAY FaOHE MASHI，hallo，bonjour，Sulut，moiẽn，hej,hallå，halló，illāc，‏هتاف للترحيب, ‏أهلا，!السلام عليكم，درود，הלו ，גוט־מאָרגן ，привет，Dzień dobry，байна уу,мэнд її，नम्स्कार，नमस्ते"
		s1 := u.EncryptAes(s, key[2:], iv[5:])
		s2 := u.DecryptAes(s1, key[2:], iv[5:])
		fmt.Println("  Test Encrypt: \033[33m", s[0:20]+"...", "\033[0m => \033[33m", s1, "\033[0m")
		fmt.Println("  Test Decrypt: \033[33m", s1, "\033[0m => \033[33m", s2[0:20]+"...", "\033[0m")
		if s2 != s {
			fmt.Println("\033[31mTest Failed\033[0m")
			fmt.Println("  \033[33m", s, "\033[0m")
			fmt.Println("  \033[33m", s2, "\033[0m")
		} else {
			fmt.Println()
			fmt.Println("\033[32mTest Succeed\033[0m")
		}

	case "-e":
		key, iv := loadKey(keyPath + os.Args[2])
		s := os.Args[3]
		s1 := u.EncryptAes(s, key[2:], iv[5:])
		s2 := u.DecryptAes(s1, key[2:], iv[5:])
		fmt.Println("Encrypted: \033[33m", s1, "\033[0m")
		if s2 != s {
			fmt.Println("\033[31mTest Failed\033[0m")
			fmt.Println("  \033[33m", s, "\033[0m")
			fmt.Println("  \033[33m", s2, "\033[0m")
		} else {
			fmt.Println()
			fmt.Println("\033[32mDecrypt test Succeed\033[0m")
		}

	case "-d":
		key, iv := loadKey(keyPath + os.Args[2])
		s2 := u.DecryptAes(os.Args[3], key[2:], iv[5:])
		fmt.Println("Decrypted: \033[33m", s2, "\033[0m")

	case "-o":
		key, iv := loadKey(keyPath + os.Args[2])
		keyOffsets := make([]int, 40)
		ivOffsets := make([]int, 40)
		for i := 0; i < 40; i++ {
			keyOffsets[i] = u.GlobalRand1.Intn(127)
			ivOffsets[i] = u.GlobalRand2.Intn(127)
			if key[i] > 127 {
				keyOffsets[i] *= -1
			}
			if iv[i] > 127 {
				ivOffsets[i] *= -1
			}
			key[i] = byte(int(key[i]) + keyOffsets[i])
			iv[i] = byte(int(iv[i]) + ivOffsets[i])
		}
		fmt.Println("package main")
		fmt.Println()
		fmt.Println("import (")
		fmt.Println("	\"fmt\"")
		fmt.Println("	\"github.com/ssgo/u\"")
		fmt.Println("	\"os\"")
		fmt.Println(")")
		fmt.Println()
		fmt.Println("func main() {")
		fmt.Println("	key := make([]byte, 0)")
		fmt.Println("	iv := make([]byte, 0)")
		fmt.Println()
		for j := 0; j < 4; j++ {
			fmt.Print("	key = append(key")
			for i := 0; i < 10; i++ {
				fmt.Print(", ", key[j*10+i])
			}
			fmt.Println(")")
		}
		fmt.Println()
		for j := 0; j < 4; j++ {
			fmt.Print("	iv = append(iv")
			for i := 0; i < 10; i++ {
				fmt.Print(", ", iv[j*10+i])
			}
			fmt.Println(")")
		}
		fmt.Println()
		for i := 39; i >= 0; i-- {
			iv[39] = byte(int(iv[39]) - 29)
			fmt.Print("	key[", i, "] = byte(int(key[", i, "]) - ", keyOffsets[i], ")\n")
			fmt.Print("	iv[", i, "] = byte(int(iv[", i, "]) - ", ivOffsets[i], ")\n")
		}
		fmt.Println()

		fmt.Println("	if len(os.Args) < 2 {")
		fmt.Println("		fmt.Println(\"need data\")")
		fmt.Println("		return")
		fmt.Println("	}")

		fmt.Println("	s1 := u.EncryptAes(os.Args[1], key[2:], iv[5:])")
		fmt.Println("	s2 := u.DecryptAes(s1, key[2:], iv[5:])")
		fmt.Println("	fmt.Println(\"Encrypted: \", s1)")
		fmt.Println("	fmt.Println(\"Decrypted check ok? \", s2 == os.Args[1])")

		fmt.Println("}")

	default:
		printUsage()
	}
	fmt.Println()
}

func scanLine(hint string) string {
	fmt.Print(hint)
	inputReader := bufio.NewReader(os.Stdin)
	line, _ := inputReader.ReadString('\n')
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[0 : len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[0 : len(line)-1]
	}
	return line
}

func loadKey(keyFile string) ([]byte, []byte) {
	fi, err := os.Stat(keyFile)
	if err != nil || fi == nil {
		fmt.Println("\033[31m", keyFile, " not exists\033[0m")
		os.Exit(0)
	}

	fd, err := os.OpenFile(keyFile, os.O_RDONLY, 0400)
	if err != nil {
		fmt.Println("\033[31mbad key file\033[0m")
		fmt.Println("\033[31m", err, "\033[0m")
		os.Exit(0)
	}

	readBuf := make([]byte, 1024)
	readSize, err := fd.Read(readBuf)
	if err != nil {
		fmt.Println("\033[31m", err, "\033[0m")
		os.Exit(0)
	}
	fd.Close()

	buf := make([]byte, 100)
	n, err := base64.StdEncoding.Decode(buf, readBuf[0:readSize])
	if err != nil {
		fmt.Println("\033[31m", err, "\033[0m")
		os.Exit(0)
	}
	if n != 81 {
		fmt.Println("\033[31mbad key length", n, "\033[0m")
		os.Exit(0)
	}
	if buf[80] != 217 {
		fmt.Println("\033[31mbad check bit", buf[80], "\033[0m")
		os.Exit(0)
	}

	return buf[0:40], buf[40:80]
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	sskey")
	fmt.Println("	\033[36m-l\033[0m		\033[37mList all saved keys\033[0m")
	fmt.Println("	\033[36m-c keyName\033[0m	\033[37mCreate a new key and save it\033[0m")
	fmt.Println("	\033[36m-t keyName\033[0m	\033[37mTest key\033[0m")
	fmt.Println("	\033[36m-e keyName data\033[0m	\033[37mEncrypt data by specified key\033[0m")
	fmt.Println("	\033[36m-d keyName data\033[0m	\033[37mDecrypt data by specified key\033[0m")
	fmt.Println("	\033[36m-o keyName\033[0m	\033[37mOutput golang code\033[0m")
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println("	\033[36msskey -l\033[0m")
	fmt.Println("	\033[36msskey -c aaa\033[0m")
	fmt.Println("	\033[36msskey -t aaa\033[0m")
	fmt.Println("	\033[36msskey -e aaa 123456\033[0m")
	fmt.Println("	\033[36msskey -d aaa gAx9Wq7YN85WKSFj7kBcHg==\033[0m")
	fmt.Println("	\033[36msskey -o aaa\033[0m")
	fmt.Println("")
}
