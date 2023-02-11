package sskeylib

import (
	"bytes"
	"errors"
	"github.com/ssgo/u"
	"text/template"
)

var codeTemplates = map[string]string{
	"go":        goTpl,
	"encryptor": encryptorTpl,
	"php":       phpTpl,
	"java":      javaTpl,
}

var goTpl = `package main

func init() {
    key := make([]byte, 0)
    iv := make([]byte, 0)

    {{range .key}}
    key = append(key, {{.}}){{end}}

    {{range .iv}}
    iv = append(iv, {{.}}){{end}}

    {{range $i, $v := .keyOffsets}}
    key[{{$i}}] = byte(int(key[{{$i}}]) - {{$v}}){{end}}

    {{range $i, $v := .ivOffsets}}
    iv[{{$i}}] = byte(int(iv[{{$i}}]) - {{$v}}){{end}}

    setSSKey(key[2:], iv[5:])
}
`

var encryptorTpl = `package main
import (
	"fmt"
	"os"
	"github.com/ssgo/u"
)

func main() {
    key := make([]byte, 0)
    iv := make([]byte, 0)

    {{range .key}}
    key = append(key, {{.}}){{end}}

    {{range .iv}}
    iv = append(iv, {{.}}){{end}}

    {{range $i, $v := .keyOffsets}}
    key[{{$i}}] = byte(int(key[{{$i}}]) - {{$v}}){{end}}

    {{range $i, $v := .ivOffsets}}
    iv[{{$i}}] = byte(int(iv[{{$i}}]) - {{$v}}){{end}}

	if len(os.Args) < 2 {
		fmt.Println("need data")
		return
	}

	s1 := u.EncryptAes(os.Args[1], key[2:], iv[5:])
	s2 := u.DecryptAes(s1, key[2:], iv[5:])

	fmt.Println("Encrypted: ", s1)
	fmt.Println("Decrypted check ok? ", s2 == os.Args[1])
}
`

var phpTpl = `<?php

$__sskeyStarer = function () {
    if(!function_exists('set_sskey')) return;
    $key = [];
    $iv = [];

    {{range .key}}
    array_push($key, {{.}});{{end}}

    {{range .iv}}
    array_push($iv, {{.}});{{end}}

    {{range $i, $v := .keyOffsets}}
    $key[{{$i}}] -= {{$v}};{{end}}

    {{range $i, $v := .ivOffsets}}
    $iv[{{$i}}] -= {{$v}};{{end}}

    set_sskey(array_slice($key, 2), array_slice($iv, 5));
};

$__sskeyStarer();
unset($__sskeyStarer);
`

var javaTpl = `
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;

public final class SSKeyStarter {
    private static boolean inited = false;
    public static void init() {
        if (inited) return;
        inited = true;

        char key[] = new char[40];
        char iv[] = new char[40];

        {{range $i, $v := .key}}
        key[{{$i}}] = {{$v}};{{end}}

        {{range $i, $v := .iv}}
        iv[{{$i}}] = {{$v}};{{end}}

        {{range $i, $v := .keyOffsets}}
        key[{{$i}}] -= {{$v}};{{end}}

        {{range $i, $v := .ivOffsets}}
        iv[{{$i}}] -= {{$v}};{{end}}

        key = Arrays.copyOfRange(key, 2, key.length);
        iv = Arrays.copyOfRange(iv, 5, iv.length);

        try {
            Class c = Class.forName("SSKeySetter");
            Method m = c.getMethod("set", byte[].class, byte[].class);
            m.invoke(null, new String(key).getBytes(StandardCharsets.ISO_8859_1), new String(iv).getBytes(StandardCharsets.ISO_8859_1));
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
`

func MakeCode(codeName string, key, iv []byte) (string, error) {
	tpl := codeTemplates[codeName]
	if tpl == "" {
		return "", errors.New("tpl not exists: " + codeName)
	}
	//fmt.Println(len(key), len(iv))
	//if len(key) < 40 || len(iv) < 40 {
	//	return "", errors.New("bad key or iv")
	//}

	keyOffsets := make([]int, len(key))
	ivOffsets := make([]int, len(iv))
	for i := 0; i < len(key); i++ {
		keyOffsets[i] = u.GlobalRand1.Intn(127)
		if key[i] > 127 {
			keyOffsets[i] *= -1
		}
		key[i] = byte(int(key[i]) + keyOffsets[i])
	}

	for i := 0; i < len(iv); i++ {
		ivOffsets[i] = u.GlobalRand2.Intn(127)
		if iv[i] > 127 {
			ivOffsets[i] *= -1
		}
		iv[i] = byte(int(iv[i]) + ivOffsets[i])
	}

	t := template.New(codeName)
	t, err := t.Parse(tpl)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, map[string]interface{}{
		"key":        key,
		"iv":         iv,
		"keyOffsets": keyOffsets,
		"ivOffsets":  ivOffsets,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
