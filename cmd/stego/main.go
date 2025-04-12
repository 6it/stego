package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/6it/stego"
)

var (
	pictureInputFile string
	messageInputFile string
	outputFile       string
	decode           bool
	encode           bool
	help             bool
)

func init() {
	flag.BoolVar(&encode, "e", false, "Encode a message to a given image file")
	flag.BoolVar(&decode, "d", false, "Decode a message from a given image file")
	flag.StringVar(&pictureInputFile, "i", "", "Path to the the input image")
	flag.StringVar(&messageInputFile, "m", "", "Path to the message input file")
	flag.StringVar(&outputFile, "o", "", "Path to the the output image/text")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()
}

func main() {
	var (
		err                      error
		picture, message, output *os.File
	)
	flag.Parse()

	if encode {
		if len(outputFile) == 0 {
			outputFile = pictureInputFile + "-out.png"
		}

		ext := outputFile[len(outputFile)-4:]
		if ext != ".png" {
			outputFile += "-out.png"
		}

		defer picture.Close()
		if picture, err = os.Open(pictureInputFile); err != nil {
			fmt.Println(err)
			return
		}

		defer message.Close()
		if message, err = os.Open(messageInputFile); err != nil {
			fmt.Println(err)
			return
		}

		defer output.Close()
		if output, err = os.Create(outputFile); err != nil {
			fmt.Println(err)
			return
		}

		if err = stego.Encode(picture, message, output); err != nil {
			fmt.Println(err)
			return
		}
	} else if decode {
		if len(outputFile) == 0 {
			outputFile = pictureInputFile + ".out.txt"
		}

		ext := outputFile[len(outputFile)-4:]
		if ext != ".txt" {
			outputFile += ".out.txt"
		}

		defer picture.Close()
		if picture, err = os.Open(pictureInputFile); err != nil {
			fmt.Println(err)
			return
		}

		defer output.Close()
		if output, err = os.Create(outputFile); err != nil {
			fmt.Println(err)
			return
		}

		if err = stego.Decode(picture, output); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println(`
stego  <https://github.com/6it/stego>
Usage: stego [command] [args]

Command:

    -h
        show this help text

    -e
        Encode a message to a given image file

    -d
        Decode a message from a given image file


Args:

    -i
        Path to the the input image

    -m
        Path to the message input file

    -o
        Path to the the output image/text
		

Example:

    encode example:
        
		stego -e -i source.jpg -m message.txt
		
		or with specified output location / file name

		stego -e -i source.jpg -m message.txt -o result.png

    decode example:

		stego -d -i source.jpg
		
		or with specified output location / file name

		stego -d -i source.jpg  -o result.txt

		`)
	}
}
