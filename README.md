# gorun

Builds and executes go packages

## What is gorun

Helps you to run your go code. It automatically sets the GoPath to the current project.
Gorun compiles the go code and copies all folders needed from the bin folder. It can run or test the code.

## Install gorun

```bash
go get github.com/starmanmaritn/gorun
```

To run gorun it is necessary to add the */bin* folder in your *GoPath* to your *Path* environment variable

## Simple build

To build your code via gorun you need to navigate to the base dir of your project were your *src, bin and pkg* folders are.
Now just enter following to your terminal:

```bash
gorun [Path to variable]
```

Example by the gorun project:

```bash
gorun github.com/starmanmaritn/gorun
```

## Execute your code

The same as before you just neet to add *-e* tag to the command

```bash
gorun -e [Path to variable]
```

### More options

 * `-w` this tag keeps gorun watching your code and rebuilds on any change
 * `-p [Folder list]` this tag copies all folders in the folder list from your *src* to your *bin*

## Run your tests

to run your tests you can add the *-t* tag to the command

```bash
gorun -t [Path to variable]
```

The only tag witch goes with the *-t* tag is the *-b* to run banchmark tests 

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