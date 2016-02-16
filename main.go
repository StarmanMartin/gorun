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
    "runtime"
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

func copyPackage(dir, packageName string) (isPackage bool, err error) {
	if len(outputString) == 0 {
		return
	}

	isPackage = true
	dest := dir + "/bin/"
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
		log.Printf("Done!! (%s)\n", runtime.GOOS)
		os.Setenv("GOPATH", currentPath)
	}()
    
    var newPath []string
    switch runtime.GOOS {
    case "windows":
        newPath = []string{newRoot, ";", currentPath}
    default: 
        newPath = []string{newRoot, ":", currentPath}
    }
	

	os.Setenv("GOPATH", strings.Join(newPath, ""))
	runBuild()
}

func runBuild() {
	buildCommandList := buildCommand(packageName)
	buildCommandList = append(buildCommandList, packageName)
	_, err := exeCmd(buildCommandList)
	if err != nil {
		log.Println(err)
		if isWatch {
			watch(nil, newRoot, false)
		}
        
		return
	}

	funcName := lastPart.FindString(packageName)
	_, err = copyPackage(newRoot, packageName)

	if err != nil {
		log.Fatal(err)
		return
	}

	if isExecute && !isTest {
		log.Printf("Running %s\n", funcName)
		executionPath := newRoot + "/bin/" + funcName
        if runtime.GOOS == "windows" {
            executionPath += ".exe"
        }
        
		exArgs := []string{executionPath}
		exArgs = append(exArgs, restArgs...)
		if isWatch {
			watch(exArgs, newRoot+"/src/"+packageName, true)
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

func watch(args []string, rootPath string, run bool) {
	doneWithoutErr := make(chan bool, 1)
	done := make(chan bool, 1)
    var cmd *exec.Cmd
    
	if run {
		cmd = getCmd(args)
		go func() {
			err := cmd.Run()
			select {
			case <-done:
				log.Println("Prozess killed")
			default:
				if err != nil {
					log.Println("process done with error = %v", err)
				}

				doneWithoutErr <- true
			}
		}()
	} else {
        doneWithoutErr <- true
    }

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

	done <- <-restart
	select {
	case <-doneWithoutErr:
		log.Println("process restarted")
		runBuild()
	default:
		log.Println("Killing prozess...")
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill: ", err)
		}

		log.Println("process restarted")
		runBuild()

	}

}
