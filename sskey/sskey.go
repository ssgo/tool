package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/tool/sskey/sskeylib"
	"github.com/ssgo/u"
	"os"
	"strconv"
	"strings"
	"time"
)

var defaultKeyIv = []byte("?GQ$0K0GgLdO=f+~L68PLm$uhKr4'=tVVFs7@sK61cj^f?HZ")
var defaultKey = defaultKeyIv[:32] //?GQ$0K0GgLdO=f+~L68PLm$uhKr4'=tV
var defaultIv = defaultKeyIv[32:]

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

	if op == "-e" || op == "-d" || op == "-e4" || op == "-d4" {
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
		files, err := os.ReadDir(keyPath)
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
			fmt.Println(u.Cyan(fileName), "	", u.White(keyPath+fileName))
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
		s1 := u.EncryptAes("Hello World!", key, iv)
		s2 := u.DecryptAes(s1, key, iv)
		fmt.Println(u.Cyan(keyName), " Created at", keyPath+keyName)
		fmt.Println("  Test Encrypt: ", u.Yellow("Hello World! "+s1))
		fmt.Println("  Test Decrypt: ", u.Yellow(s1), "=>", u.Yellow(s2))

	case "-t":
		key, iv := loadKey(keyPath + os.Args[2])
		s := "你好，Hello，안녕하세요，こんにちは，ON LI DAY FaOHE MASHI，hallo，bonjour，Sulut，moiẽn，hej,hallå，halló，illāc，‏هتاف للترحيب, ‏أهلا，!السلام عليكم，درود，הלו ，גוט־מאָרגן ，привет，Dzień dobry，байна уу,мэнд її，नम्स्कार，नमस्ते"
		s1 := u.EncryptAes(s, key, iv)
		s2 := u.DecryptAes(s1, key, iv)
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

	case "-e", "-e4":
		var key, iv []byte
		var s string
		if len(os.Args) > 3 {
			key, iv = loadKey(keyPath + os.Args[2])
			s = os.Args[3]
		} else {
			key = defaultKey
			iv = defaultIv
			s = os.Args[2]
		}
		if u.FileExists(s) {
			// 加密文件
			if in, err := u.ReadFileBytes(s); err == nil {
				outputFile := s + ".e"
				if len(os.Args) > 4 {
					outputFile = os.Args[4]
				}
				if ofp, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err == nil {
					var out []byte
					if op == "-e4" {
						out = EncryptSM4Bytes(in, key, iv)
					} else {
						aes := u.NewAes(key, iv)
						out, _ = aes.EncryptBytes(in)
					}
					_, _ = ofp.Write(out)
					_ = ofp.Close()
				}
			}
		} else {
			// 加密字符串
			var sb []byte
			if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
				bs := strings.ReplaceAll(s, " ", ",")
				//sb := make([]byte, 0)
				u.UnJson(bs, &sb)
				fmt.Println("Encrypting bytes: ", sb)
			} else {
				sb = []byte(s)
				fmt.Println("Encrypting string: ", s)
			}
			var s1, s2 string
			if op == "-e4" {
				s1 = EncryptSM4(sb, key, iv)
				s2 = string(DecryptSM4(s1, key, iv))
			} else {
				aes := u.NewAes(key, iv)
				s1 = aes.EncryptBytesToUrlBase64(sb)
				s2 = aes.DecryptUrlBase64ToString(s1)
				//s1 = u.EncryptAes(s, key, iv)
				//s2 = u.DecryptAes(s1, key, iv)
			}

			s1Bytes := u.UnUrlBase64(s1)
			fmt.Println("Encrypted base64: ", u.Yellow(u.Base64(s1Bytes)))
			fmt.Println("Encrypted url base64: ", u.Yellow(s1))
			fmt.Println("Encrypted hex: ", u.Yellow(hex.EncodeToString(s1Bytes)))
			fmt.Println("Encrypted bytes: ", s1Bytes)
			if s2 != string(sb) {
				fmt.Println(u.Red("Test Failed"))
				fmt.Println(u.Yellow(s))
				fmt.Println(u.Yellow(s2))
			} else {
				fmt.Println()
				fmt.Println(u.Green("Decrypt test Succeed"))
			}
		}

	case "-d", "-d4":
		var key, iv []byte
		var s string
		if len(os.Args) > 3 {
			key, iv = loadKey(keyPath + os.Args[2])
			s = os.Args[3]
		} else {
			key = defaultKey
			iv = defaultIv
			s = os.Args[2]
		}

		if u.FileExists(s) {
			// 解密文件
			if in, err := u.ReadFileBytes(s); err == nil {
				outputFile := s + ".d"
				if strings.HasSuffix(s, ".e") {
					outputFile = s[0 : len(s)-2]
				}
				if len(os.Args) > 4 {
					outputFile = os.Args[4]
				}
				if ofp, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err == nil {
					var out []byte
					if op == "-d4" {
						out = DecryptSM4Bytes(in, key, iv)
					} else {
						aes := u.NewAes(key, iv)
						out, _ = aes.DecryptBytes(in)
					}
					_, _ = ofp.Write(out)
					_ = ofp.Close()
				}
			}
		} else {
			// 解密字符串
			var b2 []byte
			var err error
			if op == "-d4" {
				b2 = DecryptSM4Bytes(u.UnUrlBase64(s), key, iv)
			} else {
				b2, err = u.DecryptAesBytes(u.UnUrlBase64(s), key, iv)
			}
			if err != nil || b2 == nil || len(b2) == 0 {
				if op == "-d4" {
					b2 = DecryptSM4Bytes(u.UnBase64(s), key, iv)
				} else {
					b2, err = u.DecryptAesBytes(u.UnBase64(s), key, iv)
				}
				if err != nil || b2 == nil || len(b2) == 0 {
					b2e, _ := hex.DecodeString(s)
					if op == "-d4" {
						b2 = DecryptSM4Bytes(b2e, key, iv)
					} else {
						b2, err = u.DecryptAesBytes(b2e, key, iv)
					}
					if err != nil {
						fmt.Println(u.Red(err.Error()))
					}
				}
			}

			fmt.Println("Decrypted: ", u.Yellow(string(b2)))
			fmt.Println("Decrypted bytes: ", b2)
		}
	case "-php":
		makeCode("php", keyPath)
	case "-java":
		makeCode("java", keyPath)
	case "-go":
		makeCode("go", keyPath)
	case "-o":
		var useKey, useIv, forKey, forIv []byte
		if len(os.Args) > 3 {
			useKey, useIv = loadKey(keyPath + os.Args[2])
			forKey, forIv = loadKey(keyPath + os.Args[3])
			fmt.Println("Encrypted key: ", u.Yellow(u.EncryptAes(string(forKey), useKey, useIv)))
			fmt.Println("Encrypted key bytes: ", u.UnUrlBase64(u.EncryptAes(string(forKey), useKey, useIv)))
			fmt.Println("Encrypted iv: ", u.Yellow(u.EncryptAes(string(forIv), useKey, useIv)))
			fmt.Println("Encrypted iv bytes: ", u.UnUrlBase64(u.EncryptAes(string(forIv), useKey, useIv)))
		} else {
			forKey, forIv = loadKey(keyPath + os.Args[2])
			fmt.Println("key base64: ", u.Yellow(u.Base64(forKey)))
			fmt.Println("key url base64: ", u.Yellow(u.UrlBase64(forKey)))
			fmt.Println("key hex: ", u.Yellow(hex.EncodeToString(forKey)))
			fmt.Println("key bytes: ", forKey)
			fmt.Println("iv base64: ", u.Yellow(u.Base64(forIv)))
			fmt.Println("iv url base64: ", u.Yellow(u.UrlBase64(forIv)))
			fmt.Println("iv hex: ", u.Yellow(hex.EncodeToString(forIv)))
			fmt.Println("iv bytes: ", forIv)
		}

	case "-sync":
		syncSSKeys(keyPath)
	default:
		printUsage()
	}
	fmt.Println()
}

func syncSSKeys(keyPath string) {
	lenArgs := len(os.Args)
	if lenArgs < 3 {
		fmt.Println("please enter your key name!")
		return
	}
	if lenArgs < 4 {
		fmt.Println("please enter your upload url!")
		return
	}
	keyNames := strings.Split(os.Args[2], ",")
	var encryptedKeys = map[string]string{}
	var settedKey []byte
	var settedIv []byte
	//use sync key
	var settedKeyIv = getKey(keyPath+"sync", true)
	if bytes.Equal(settedKeyIv, defaultKeyIv) {
		settedKey = defaultKey
		settedIv = defaultIv
	} else {
		settedKey = settedKeyIv[2:40]
		settedIv = settedKeyIv[45:]
	}
	for _, keyName := range keyNames {
		keyName = strings.Trim(keyName, " ")
		if len(keyName) < 1 {
			fmt.Println("invalid key name")
			return
		}
		encryptedKeys[keyName] = u.EncryptAes(string(getKey(keyPath+keyName, false)[:80]), settedKey, settedIv)
	}
	sendKeys := httpclient.GetClient(10*time.Second).Post(os.Args[3], encryptedKeys)
	if sendKeys.Error != nil {
		fmt.Println("Error ", sendKeys.Error)
		return
	}
	if sendKeys.String() != "true" {
		fmt.Println("sync keys failed")
		return
	}
	fmt.Println("send keys detail:")
	fmt.Println(encryptedKeys)
	fmt.Println("send keys successfully")
}

func makeCode(codeName string, keyPath string) {
	lenArgs := len(os.Args)
	if lenArgs < 3 {
		fmt.Println("please enter your key name!")
		return
	}
	buf := getKey(keyPath+os.Args[2], false)
	codeDetail, err := sskeylib.MakeCode(codeName, buf[0:40], buf[40:80])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(codeDetail)
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

func getKey(keyFile string, usedDefault bool) []byte {
	fi, err := os.Stat(keyFile)
	if err != nil || fi == nil {
		if usedDefault {
			return defaultKeyIv
		}
		fmt.Println(u.Red(keyFile), u.Red("does not exists"))
		os.Exit(0)
	}

	fd, err := os.OpenFile(keyFile, os.O_RDONLY, 0400)
	if err != nil {
		if usedDefault {
			return defaultKeyIv
		}
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
	return buf
}

func loadKey(keyFile string) ([]byte, []byte) {
	buf := getKey(keyFile, false)
	key := make([]byte, 40)
	iv := make([]byte, 40)
	for i := 0; i < 40; i++ {
		key[i] = buf[i]
		iv[i] = buf[40+i]
	}
	return key[2:], iv[5:]
}

func printUsage() {
	fmt.Println("sskey.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("	sskey")
	fmt.Println(u.Cyan("	-l		") + u.White("List all saved keys"))
	fmt.Println(u.Cyan("	-c keyName	") + u.White("Create a new key and save it"))
	fmt.Println(u.Cyan("	-t keyName	") + u.White("Test key"))
	fmt.Println(u.Cyan("	-e [keyName] data	") + u.White("Encrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-d [keyName] data	") + u.White("Decrypt data by specified key or default key"))
	fmt.Println(u.Cyan("	-e4 [keyName] data	") + u.White("Encrypt data by specified key or default key with SM4"))
	fmt.Println(u.Cyan("	-d4 [keyName] data	") + u.White("Decrypt data by specified key or default key with SM4"))
	fmt.Println(u.Cyan("	-php keyName	") + u.White("Output php code"))
	fmt.Println(u.Cyan("	-java keyName	") + u.White("Output java code"))
	fmt.Println(u.Cyan("	-go keyName	") + u.White("Output go code"))
	fmt.Println(u.Cyan("	-o keyName	") + u.White("Output key&iv by default key"))
	fmt.Println(u.Cyan("	-o [byKeyName] keyName	") + u.White("Output key&iv by specified key)"))
	fmt.Println(u.Cyan("	-sync keyNames	") + u.White("Synchronization of keys to another machine from url"))
	fmt.Println("")
	fmt.Println("Samples:")
	fmt.Println(u.Cyan("	sskey -l"))
	fmt.Println(u.Cyan("	sskey -c aaa"))
	fmt.Println(u.Cyan("	sskey -t aaa"))
	fmt.Println(u.Cyan("	sskey -e 123456"))
	fmt.Println(u.Cyan("	sskey -d xxxxxx"))
	fmt.Println(u.Cyan("	sskey -e aaa 123456"))
	fmt.Println(u.Cyan("	sskey -d aaa xxxxxx"))
	fmt.Println(u.Cyan("	sskey -php aaa"))
	fmt.Println(u.Cyan("	sskey -java aaa"))
	fmt.Println(u.Cyan("	sskey -go aaa"))
	fmt.Println(u.Cyan("	sskey -o aaa"))
	fmt.Println(u.Cyan("	sskey -o bbb aaa"))
	fmt.Println(u.Cyan("	sskey -sync aaa,bbb,ccc http://xxxxxx"))
	fmt.Println("")
}
