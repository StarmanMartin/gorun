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
	"time"

	"github.com/starmanmartin/simple-fs"
)

const (
	install = "install"
	test    = "test"
)

var runTypes = []string{install, test}

var (
	lastPart                                        *regexp.Regexp
	isTest, isBenchTest, isExecute, isWatch         bool
	newRoot, packageName, currentPath, outputString string
	restArgs                                        []string
)

func init() {
	lastPart, _ = regexp.Compile(`[^\\/]*$`)

	flag.BoolVar(&isTest, "t", false, "Run as Test")
	flag.BoolVar(&isBenchTest, "b", false, "Bench tests (only if test)")
	flag.BoolVar(&isExecute, "e", false, "Execute (only if not test)")
	flag.BoolVar(&isWatch, "w", false, "Execute (only if not test)")
	flag.StringVar(&outputString, "p", "", "Make Package")
}

func getCmd(cmdCommand []string) *exec.Cmd {
	parts := cmdCommand
	head := parts[0]
	parts = parts[1:len(parts)]

	cmd := exec.Command(head, parts...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func exeCmd(cmdCommand []string) (*exec.Cmd, error) {
	cmd := getCmd(cmdCommand)

	if err := cmd.Run(); err != nil {
		return cmd, err
	}

	return cmd, nil
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

func copyPackage(dir, packageName, funcName string) (isPackage bool, err error) {
	if len(outputString) == 0 {
		return
	}

	isPackage = true
	dest := dir + "/bin/" + funcName + "/"
	src := dir + "/src/" + packageName + "/"

	output := strings.Split(outputString, " ")

	for _, dirName := range output {
		err = fs.CopyFolder(src+dirName, dest+dirName)
		if err != nil {
			return
		}
	}

	return
}

func main() {
	flag.Parse()
	var err error
	newRoot, packageName, restArgs, err = handelPathArgs()
	if err != nil {
		log.Fatal(err)
		return
	}
	currentPath := os.Getenv("GOPATH")
	defer func() {
		log.Println("Done!!")
		os.Setenv("GOPATH", currentPath)
	}()
	newPath := []string{newRoot, ";", currentPath}

	os.Setenv("GOPATH", strings.Join(newPath, ""))
	runBuild()
}

func runBuild() {
	buildCommandList := buildCommand(packageName)
	buildCommandList = append(buildCommandList, packageName)
    _, err := exeCmd(buildCommandList)
	if err != nil {
		log.Fatal(err)
	}

	funcName := lastPart.FindString(packageName)
	isPackage, err := copyPackage(newRoot, packageName, funcName)

	if err != nil {
		log.Fatal(err)
		return
	} else if isPackage {
		fs.SyncFile(newRoot+"/bin/"+funcName+".exe", newRoot+"/bin/"+funcName+"/"+funcName+".exe")
		funcName = funcName + "/" + funcName
	}

	if isExecute && !isTest {
		log.Printf("Running %s\n", funcName)
		executionPath := newRoot + "/bin/" + funcName + ".exe"
		exArgs := []string{executionPath}
		exArgs = append(exArgs, restArgs...)
		if isWatch {
			watch(exArgs, newRoot+"/src/"+packageName)
		} else {
			_, err := exeCmd(exArgs)
			if err != nil {
				log.Fatal(err)
				return
			}
		}
	} else {
		log.Printf("Builded %s\n", funcName)
	}
}

func watch(args []string, rootPath string) {
	done := make(chan error, 1)
	doneWithoutErr := make(chan bool, 1)

	cmd := getCmd(args)

	go func() {
		err := cmd.Run()
		if err != nil {
			done <- err
		} else {
			doneWithoutErr <- true
		}
	}()

	restart := make(chan bool, 1)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		lastChaek := time.Now()
		for _ = range ticker.C {
			isUpdated, _ := fs.CheckIfFolderUpdated(rootPath, lastChaek)
			if isUpdated {
				restart <- true
				ticker.Stop()
			}
		}
	}()

	select {
	case <-restart:
		select {
		case <-doneWithoutErr:
			log.Println("process restarted")
			runBuild()
		default:
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}

			log.Println("process restarted")
			runBuild()

		}

	case err := <-done:
		if err != nil {
			log.Fatal("process done with error = %v", err)
		} else {
			log.Print("process done gracefully without error")
		}
	}

}
