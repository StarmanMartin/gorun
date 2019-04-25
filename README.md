# gorun
[Simple build](#simple-build)

Builds and executes go packages

## What is gorun

Helps to run your go code. It automatically sets the GoPath to the current project.
Gorun compiles the go code and copies all folders to a output folder. It also can run tests.

## Install gorun

```bash
go get github.com/starmanmaritn/gorun
```

To run gorun it is necessary to add the */bin* folder of your GoPath to your *Path* environment variable

## Simple_build

To build your code via gorun you need to navigate to the base path of your current project. There should be a *src*, a *bin* and a *pkg* folder located. Now just enter following to your terminal to build your Project:

```bash
gorun [Path to variable]
```

Example for the gorun project:

```bash
gorun github.com/starmanmaritn/gorun
```

## Execute your code

The same as before, you just neet to add *-e* tag to the command

```bash
gorun -e [Path to variable]
```

### More options

 * `-w` this tag keeps gorun watching your code and rebuilds if it changes.
 * `-p [Folder list]` this tag copies all folders in the folder list from your *src* to your *bin*

## Run your tests

to run your tests you can add the *-t* tag to the command

```bash
gorun -t [Path to variable]
```

The only other tag witch goes with the *-t* tag is the *-b* to run banchmark tests 

```bash
gorun -t -b [Path to variable]
```

## VSCode task.json

Sample from [Codon-server](https://github.com/StarmanMartin/codon-server)

```json
{
	"version": "0.1.0",

	// The command is tsc. Assumes that tsc has been installed using npm install -g typescript
	"command": "gorun",

	// The command is a shell script
	"isShellCommand": true,

	// Show the output window only if unrecognized errors occur.
	"showOutput": "silent",

	// args is the HelloWorld program to compile.
	"args": ["-w", "-e", "-p", "public views config", "github.com/starmanmartin/codon-server"],

}
```
