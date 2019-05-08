package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/ssgo/u"
	"io/ioutil"
	"os"
	"strconv"
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
	if (op == "-c" || op == "-t" || op == "-o" || op == "-db" || op == "-redis") && len(os.Args) < 3 {
		keyName := scanLine(u.Cyan("Please enter key name: "))
		if keyName == "" {
			printUsage()
			fmt.Println(u.Red("need key"))
			return
		}
		os.Args = append(os.Args, keyName)
	}

	if (op == "-e" || op == "-d") && len(os.Args) < 3 {
		data := scanLine(u.Cyan("Please enter data: "))
		if data == "" {
			printUsage()
			fmt.Println(u.Red("need data"))
			return
		}
		os.Args = append(os.Args, data)
	}

	switch op {
	case "-l":
		files, err := ioutil.ReadDir(keyPath)
		if err != nil {
			fmt.Println(u.Red(err.Error()))
			return
		}
		n := 0
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' {
				continue
			}
			n++
			fmt.Println(u.Cyan(fileName), "	", u.White(keyPath+" "+fileName))
		}
		fmt.Println(n, "Keys")
	case "-c":
		keyName := os.Args[2]

		fi, err := os.Stat(keyPath + keyName)
		if err == nil && fi != nil {
			fmt.Println(u.Red("key exists"))
			return
		}

		fd, err := os.OpenFile(keyPath+keyName, os.O_CREATE|os.O_WRONLY, 0400)
		if err != nil {
			fmt.Println(u.Red("bad key file"))
			fmt.Println(u.Red(err.Error()))
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
		fmt.Println(u.Cyan(keyName), " Created at", keyPath+keyName)
		fmt.Println("  Test Encrypt: ", u.Yellow("Hello World! "+s1))
		fmt.Println("  Test Decrypt: ", u.Yellow(s1), "=>", u.Yellow(s2))

	case "-t":
		key, iv := loadKey(keyPath + os.Args[2])
		s := "你好，Hello，안녕하세요，こんにちは，ON LI DAY FaOHE MASHI，hallo，bonjour，Sulut，moiẽn，hej,hallå，halló，illāc，‏هتاف للترحيب, ‏أهلا，!السلام عليكم，درود，הלו ，גוט־מאָרגן ，привет，Dzień dobry，байна уу,мэнд її，नम्स्कार，नमस्ते"
		s1 := u.EncryptAes(s, key[2:], iv[5:])
		s2 := u.DecryptAes(s1, key[2:], iv[5:])
		fmt.Println("  Test Encrypt: ", u.Yellow(s[0:20]+"..."), "=>", u.Yellow(s1))
		fmt.Println("  Test Decrypt: ", u.Yellow(s1), "=>", u.Yellow(s2[0:20]+"..."))
		if s2 != s {
			fmt.Println(u.Red("Test Failed"))
			fmt.Println("  ", u.Yellow(s))
			fmt.Println("  ", u.Yellow(s2))
		} else {
			fmt.Println()
			fmt.Println(u.Green("Test Succeed"))
		}

	case "-e":
		var key, iv []byte
		var s string
		if len(os.Args) > 3 {
			key, iv = loadKey(keyPath + os.Args[2])
			key = key[2:]
			iv = iv[5:]
			s = os.Args[3]
		} else {
			key = []byte("?GQ$0K0GgLdO=f+~L68PLm$uhKr4'=tV")
			iv = []byte("VFs7@sK61cj^f?HZ")
			s = os.Args[2]
		}
		s1 := u.EncryptAes(s, key, iv)
		s2 := u.DecryptAes(s1, key, iv)
		fmt.Println("Encrypted: ", u.Yellow(s1))
		if s2 != s {
			fmt.Println(u.Red("Test Failed"))
			fmt.Println(u.Yellow(s))
			fmt.Println(u.Yellow(s2))
		} else {
			fmt.Println()
			fmt.Println(u.Green("Decrypt test Succeed"))
		}
	case "-d":
		var key, iv []byte
		var s string
		if len(os.Args) > 3 {
			key, iv = loadKey(keyPath + os.Args[2])
			key = key[2:]
			iv = iv[5:]
			s = os.Args[3]
		} else {
			key = []byte("?GQ$0K0GgLdO=f+~L68PLm$uhKr4'=tV")
			iv = []byte("VFs7@sK61cj^f?HZ")
			s = os.Args[2]
		}
		s2 := u.DecryptAes(s, key, iv)
		fmt.Println("Decrypted: ", u.Yellow(s2))
	case "-o":
		makeGoCode(keyPath, "redis", false)
	case "-db":
		makeGoCode(keyPath, "db", true)
	case "-redis":
		makeGoCode(keyPath, "redis", true)
	default:
		printUsage()
	}
	fmt.Println()
}

func makeGoCode(keyPath string, dataType string, forApp bool) {
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
	if dataType == "redis" {
		fmt.Println("	\"github.com/ssgo/redis\"")
	} else if dataType == "db" {
		fmt.Println("	\"github.com/ssgo/db\"")
	} else {
		fmt.Println("	\"github.com/ssgo/db\"")
	}
	fmt.Println("	\"os\"")
	fmt.Println(")")
	fmt.Println()
	if forApp {
		fmt.Println("func init() {")
	} else {
		fmt.Println("func main() {")
	}
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
	if !forApp {
		fmt.Println("	if len(os.Args) < 2 {")
		fmt.Println("		fmt.Println(\"need data\")")
		fmt.Println("		return")
		fmt.Println("	}")
	}
	if dataType == "redis" {
		fmt.Println("	redis.SetEncryptKeys(key[2:], iv[5:])")
	} else if dataType == "db" {
		fmt.Println("	db.SetEncryptKeys(key[2:], iv[5:])")
	} else {
		fmt.Println("	db.SetEncryptKeys(key[2:], iv[5:])")
	}
	//fmt.Println("	s1 := u.EncryptAes(os.Args[1], key[2:], iv[5:])")
	//fmt.Println("	s2 := u.DecryptAes(s1, key[2:], iv[5:])")
	//fmt.Println("	fmt.Println(\"Encrypted: \", s1)")
	//fmt.Println("	fmt.Println(\"Decrypted check ok? \", s2 == os.Args[1])")

	fmt.Println("}")
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
		fmt.Println(u.Red(keyFile))
		os.Exit(0)
	}

	fd, err := os.OpenFile(keyFile, os.O_RDONLY, 0400)
	if err != nil {
		fmt.Println(u.Red("bad key file"))
		fmt.Println(u.Red(err.Error()))
		os.Exit(0)
	}

	readBuf := make([]byte, 1024)
	readSize, err := fd.Read(readBuf)
	if err != nil {
		fmt.Println(u.Red(err.Error()))
		os.Exit(0)
	}
	fd.Close()

	buf := make([]byte, 100)
	n, err := base64.StdEncoding.Decode(buf, readBuf[0:readSize])
	if err != nil {
		fmt.Println(u.Red(err.Error()))
		os.Exit(0)
	}
	if n != 81 {
		fmt.Println(u.Red("bad key length " + strconv.Itoa(n)))
		os.Exit(0)
	}
	if buf[80] != 217 {
		fmt.Println(u.Red("bad check bit " + string(buf[80])))
		os.Exit(0)
	}

	return buf[0:40], buf[40:80]
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	sskey")
	fmt.Println(u.Cyan("	-l		") + u.White("List all saved keys"))
	fmt.Println(u.Cyan("	-c keyName	") + u.White("Create a new key and save it"))
	fmt.Println(u.Cyan("	-t keyName	") + u.White("Test key"))
	fmt.Println(u.Cyan("	-e [keyName] data	") + u.White("Encrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-d [keyName] data	") + u.White("Decrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-o keyName	") + u.White("Output golang code"))
	fmt.Println(u.Cyan("	-db keyName	") + u.White("db configure Output golang code"))
	fmt.Println(u.Cyan("	-redis keyName	") + u.White("redis configure Output golang code"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println(u.Cyan("	sskey -l"))
	fmt.Println(u.Cyan("	sskey -c aaa"))
	fmt.Println(u.Cyan("	sskey -t aaa"))
	fmt.Println(u.Cyan("	sskey -e 123456"))
	fmt.Println(u.Cyan("	sskey -d vcg9B/GX3Tqf1EWfpfDeMw=="))
	fmt.Println(u.Cyan("	sskey -e aaa 123456"))
	fmt.Println(u.Cyan("	sskey -d aaa gAx9Wq7YN85WKSFj7kBcHg=="))
	fmt.Println(u.Cyan("	sskey -o aaa"))
	fmt.Println(u.Cyan("	sskey -db aaa"))
	fmt.Println(u.Cyan("	sskey -redis aaa"))
	fmt.Println("")
}
