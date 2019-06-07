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
	_ = os.Mkdir(keyPath, 0700)

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

	if op == "-e" || op == "-d" {
		if len(os.Args) < 3 || (len(os.Args) == 3 && u.FileExists(keyPath+os.Args[2])) {
			data := scanLine(u.Cyan("Please enter data: "))
			if data == "" {
				printUsage()
				fmt.Println(u.Red("need data"))
				return
			}
			os.Args = append(os.Args, data)
		}
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
		_, _ = fd.WriteString(base64.StdEncoding.EncodeToString(buf))
		_ = fd.Close()

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
	case "-php":
		makePHPCode(keyPath)
	case "-java":
		makeJavaCode(keyPath)
	case "-go":
		makeGoCode(keyPath, "project")
	case "-o":
		makeGoCode(keyPath, "ops")
	default:
		printUsage()
	}
	fmt.Println()
}

func makeGoCode(keyPath string, usedType string) {
	lenArgs := len(os.Args)
	if lenArgs < 3 {
		fmt.Println("please enter your sskey file!")
		return
	}
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
	if usedType == "ops" {
		fmt.Println("import (")
		fmt.Println("	\"fmt\"")
		fmt.Println("	\"os\"")
		fmt.Println("	\"github.com/ssgo/u\"")
		fmt.Println(")")
		fmt.Println("func main() {")
	} else {
		fmt.Println("func init() {")
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

	if usedType == "ops" {
		fmt.Println("	if len(os.Args) < 2 {")
		fmt.Println("		fmt.Println(\"need data\")")
		fmt.Println("		return")
		fmt.Println("	}")
		fmt.Println("	s1 := u.EncryptAes(os.Args[1], key[2:], iv[5:])")
		fmt.Println("	s2 := u.DecryptAes(s1, key[2:], iv[5:])")
		fmt.Println("	fmt.Println(\"Encrypted: \", s1)")
		fmt.Println("	fmt.Println(\"Decrypted check ok? \", s2 == os.Args[1])")
	} else {
		fmt.Println("	setSSKey(key[2:], iv[5:])")
	}

	fmt.Println("}")
}

func makePHPCode(keyPath string) {
	lenArgs := len(os.Args)
	if lenArgs < 3 {
		fmt.Println("please enter your sskey file!")
		return
	}
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
	fmt.Println("<?php")
	fmt.Println()
	fmt.Println("$sskeyStarer = function () {")

	fmt.Println("	if(!function_exists('set_sskey')) {return;}")
	fmt.Println("	$key = [];")
	fmt.Println("	$iv = [];")
	fmt.Println()
	for j := 0; j < 4; j++ {
		fmt.Print("	array_push($key")
		for i := 0; i < 10; i++ {
			fmt.Print(", ", key[j*10+i])
		}
		fmt.Println(");")
	}
	fmt.Println()
	for j := 0; j < 4; j++ {
		fmt.Print("	array_push($iv")
		for i := 0; i < 10; i++ {
			fmt.Print(", ", iv[j*10+i])
		}
		fmt.Println(");")
	}
	fmt.Println()
	for i := 39; i >= 0; i-- {
		iv[39] = byte(int(iv[39]) - 29)
		fmt.Println("	$key[", i, "] = $key[", i, "] - ", keyOffsets[i], ";")
		fmt.Println("	$iv[", i, "] = $iv[", i, "] - ", ivOffsets[i], ";")
	}
	fmt.Println()
	fmt.Println("	$key = array_slice($key, 2);")
	fmt.Println("	$iv = array_slice($iv, 5);")
	fmt.Println("	set_sskey($key, $iv);")
	fmt.Println("};")
	fmt.Println("$sskeyStarer();")
	fmt.Println("unset($sskeyStarer);")
}

func makeJavaCode(keyPath string) {
	lenArgs := len(os.Args)
	if lenArgs < 3 {
		fmt.Println("please enter your sskey file!")
		return
	}
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
	fmt.Println("import java.lang.reflect.InvocationTargetException;")
	fmt.Println("import java.lang.reflect.Method;")
	fmt.Println("import java.util.Arrays;")
	fmt.Println()
	fmt.Println("public class SSKeyStarter {")
	fmt.Println("    private static final SSKeyStarter instance = new SSKeyStarter();")
	fmt.Println("    private SSKeyStarter() {")

	fmt.Println("        int[] key = {")

	for j := 0; j < 4; j++ {
		fmt.Print("                ")
		for i := 0; i < 10; i++ {
			fmt.Print(key[j*10+i], ", ")
		}
		fmt.Println()
	}
	fmt.Println("        };")
	fmt.Println()
	fmt.Println("        int[] iv = {")
	for j := 0; j < 4; j++ {
		fmt.Print("                ")
		for i := 0; i < 10; i++ {
			fmt.Print(iv[j*10+i], ", ")
		}
		fmt.Println()
	}
	fmt.Println("        };")
	fmt.Println()
	for i := 39; i >= 0; i-- {
		iv[39] = byte(int(iv[39]) - 29)
		fmt.Println("        key[", i, "] = key[", i, "] - ", keyOffsets[i], ";")
		fmt.Println("        iv[", i, "] = iv[", i, "] - ", ivOffsets[i], ";")
	}
	fmt.Println()
	fmt.Println("        key = Arrays.copyOfRange(key, 2, key.length);")
	fmt.Println("        iv = Arrays.copyOfRange(iv, 5, iv.length);")
	fmt.Println("        try{")
	fmt.Println("            Class c = Class.forName(\"SSKeySetter\");")
	fmt.Println("            Method m = c.getMethod(\"set\", key.getClass(),iv.getClass());")
	fmt.Println("            m.invoke(c.newInstance(), new Object[]{key, iv});")
	fmt.Println("        } catch (ClassNotFoundException e) {")
	fmt.Println("            e.printStackTrace();")
	fmt.Println("        } catch (NoSuchMethodException e) {")
	fmt.Println("            e.printStackTrace();")
	fmt.Println("        } catch (IllegalAccessException e) {")
	fmt.Println("            e.printStackTrace();")
	fmt.Println("        } catch (InvocationTargetException e) {")
	fmt.Println("            e.printStackTrace();")
	fmt.Println("        } catch (Exception e) {")
	fmt.Println("            e.printStackTrace();")
	fmt.Println("        }")
	fmt.Println("    }")
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
	_ = fd.Close()

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
	fmt.Println("Wellcome to use sskey.")
	fmt.Println("Usage:")
	fmt.Println("	sskey")
	fmt.Println(u.Cyan("	-l		") + u.White("List all saved keys"))
	fmt.Println(u.Cyan("	-c keyName	") + u.White("Create a new key and save it"))
	fmt.Println(u.Cyan("	-t keyName	") + u.White("Test key"))
	fmt.Println(u.Cyan("	-e [keyName] data	") + u.White("Encrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-d [keyName] data	") + u.White("Decrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-php keyName	") + u.White("Output php code"))
	fmt.Println(u.Cyan("	-java keyName	") + u.White("Output java code"))
	fmt.Println(u.Cyan("	-go keyName	") + u.White("Output go code"))
	fmt.Println(u.Cyan("	-o keyName	") + u.White("Encrypt tool(make executable file)"))
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
	fmt.Println(u.Cyan("	sskey -php aaa"))
	fmt.Println(u.Cyan("	sskey -java aaa"))
	fmt.Println(u.Cyan("	sskey -go aaa"))
	fmt.Println(u.Cyan("	sskey -o aaa"))
	fmt.Println("")
}
