package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/starmanmartin/simple-fs"
)

const (
	install = "install"
	test    = "test"
)

var lastPart *regexp.Regexp

var runTypes = []string{install, test}

var isTest bool
var isBenchTest bool
var isExecute bool
var currentPath string
var outputString string

func init() {
	lastPart, _ = regexp.Compile(`[^\\/]*$`)
		
	flag.BoolVar(&isTest, "t", false, "Run as Test")
	flag.BoolVar(&isBenchTest, "b", false, "Bench tests (only if test)")
	flag.BoolVar(&isExecute, "e", false, "Execute (only if not test)")
	flag.StringVar(&outputString, "p", "", "Make Package")
}

func exeCmd(cmdCommand []string) error {
	parts := cmdCommand
	head := parts[0]
	parts = parts[1:len(parts)]

	cmd := exec.Command(head, parts...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func buildCommand(packageName string) []string {
	buffer := make([]string, 0, 6)

	buffer = append(buffer, "go")

	if isTest {
		buffer = append(buffer, "test")
		if isBenchTest {
			buffer = append(buffer, "-bench=.")
		}
	} else {
		buffer = append(buffer, "install")
	}

	buffer = append(buffer, "-v")

	return buffer
}

func handelPathArgs() (string, string, []string, error) {
	args := flag.Args()
	if len(args) == 0 {
		return "", "", nil, errors.New("No Args")
	}

	if len(args) == 1 || args[0][:11] == "github.com/" {
		dir, err := os.Getwd()
		if err != nil {
			return "", "", nil, (err)
		}

		return dir, args[0], args[1:], nil
	}

	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return "", "", nil, (err)
	}

	return absPath, args[1], args[1:], nil
}

func copyPackage(dir, packageName, funcName string) (isPackage bool, err error){
	if len(outputString) == 0 {
		return
	}
	
	isPackage = true	
	dest := dir + "/bin/" + funcName + "/"
	src := dir + "/src/" + packageName + "/" 
	
	output := strings.Split(outputString, " ")
	
	for _, dirName := range output {
		err = fs.CopyFolder(src + dirName, dest + dirName)
		if err != nil {
			return
		}
	}
	
	return
}

func main() {
	flag.Parse()
	newRoot, packageName, restArgs, err := handelPathArgs()
	if err != nil {
		log.Fatal(err)
		return
	}

	buildCommandList := buildCommand(packageName)
	buildCommandList = append(buildCommandList, packageName)
	currentPath := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", currentPath)
	newPath := []string{newRoot, ";", currentPath}

	os.Setenv("GOPATH", strings.Join(newPath, ""))
	if err = exeCmd(buildCommandList); err != nil {
		log.Fatal(err)
	}

	funcName := lastPart.FindString(packageName)
	isPackage, err := copyPackage(newRoot, packageName, funcName)

	if err != nil {
		log.Fatal(err)
		return
	} else if(isPackage) {
		fs.SyncFile(newRoot + "/bin/" + funcName + ".exe", newRoot + "/bin/" + funcName + "/" + funcName + ".exe")
		funcName = funcName + "/" + funcName
	}

	if isExecute && !isTest {
		executionPath := newRoot + "/bin/" + funcName + ".exe"
		exArgs := []string{executionPath}
		exArgs = append(exArgs, restArgs...)
		if err = exeCmd(exArgs); err != nil {
			log.Fatal(err)
		}
	}
}
