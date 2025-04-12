package stego

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
)

// r (image from file), message (text file), w (output writer)
func Encode(r io.Reader, message io.Reader, w io.Writer) (err error) {
	var (
		rgbImg     *image.NRGBA
		pixColor   color.NRGBA
		bitFromMsg byte
		ok         bool
		msg        []byte
	)

	if rgbImg, err = toNRGBA(r); err != nil {
		return
	}

	var (
		width         = rgbImg.Bounds().Dx()
		height        = rgbImg.Bounds().Dy()
		maxEncodeSize = ((width * height * 3) / 8)
		ch            = make(chan byte, 100)
	)

	if msg, err = io.ReadAll(message); err != nil {
		return
	}

	if maxEncodeSize < (len(msg) + 4) {
		return fmt.Errorf("this image only can hold %v byte message in length", maxEncodeSize-4)
	}

	setHeader(&msg, uint32(len(msg)))

	go getEachBitMessage(msg, ch)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixColor = rgbImg.NRGBAAt(x, y)

			// red
			bitFromMsg, ok = <-ch
			if !ok {
				rgbImg.SetNRGBA(x, y, pixColor)
				break
			}
			setLSB(&pixColor.R, bitFromMsg)

			// green
			bitFromMsg, ok = <-ch
			if !ok {
				rgbImg.SetNRGBA(x, y, pixColor)
				break
			}
			setLSB(&pixColor.G, bitFromMsg)

			// blue
			bitFromMsg, ok = <-ch
			if !ok {
				rgbImg.SetNRGBA(x, y, pixColor)
				break
			}
			setLSB(&pixColor.B, bitFromMsg)

			rgbImg.SetNRGBA(x, y, pixColor)
		}
	}

	return png.Encode(w, rgbImg)
}

// r (input image file to decode), w (output text file)
func Decode(r io.Reader, w io.Writer) (err error) {
	var (
		rgbImg *image.NRGBA
	)

	if rgbImg, err = toNRGBA(r); err != nil {
		return
	}

	length := getMessageLength(rgbImg)
	message := decodeNRGBA(4, length, rgbImg)

	if _, err = w.Write(message); err != nil {
		return
	}
	return
}

func decodeNRGBA(startOffset uint32, msgLen uint32, rgbImg *image.NRGBA) (msg []byte) {
	var (
		byteIndex uint32
		bitIndex  uint32
		c         color.NRGBA
		lsb       byte
		width     = rgbImg.Bounds().Dx()
		height    = rgbImg.Bounds().Dy()
	)

	msg = append(msg, 0)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c = rgbImg.NRGBAAt(x, y)

			// red
			lsb = getLSB(c.R)
			msg[byteIndex] = setBitInByte(msg[byteIndex], bitIndex, lsb)
			bitIndex++

			if bitIndex > 7 { // to the next byte
				bitIndex = 0
				byteIndex++

				if byteIndex >= msgLen+startOffset {
					return msg[startOffset : msgLen+startOffset]
				}

				msg = append(msg, 0)
			}

			// green
			lsb = getLSB(c.G)
			msg[byteIndex] = setBitInByte(msg[byteIndex], bitIndex, lsb)
			bitIndex++

			if bitIndex > 7 {

				bitIndex = 0
				byteIndex++

				if byteIndex >= msgLen+startOffset {
					return msg[startOffset : msgLen+startOffset]
				}

				msg = append(msg, 0)
			}

			//blue
			lsb = getLSB(c.B)
			msg[byteIndex] = setBitInByte(msg[byteIndex], bitIndex, lsb)
			bitIndex++

			if bitIndex > 7 {
				bitIndex = 0
				byteIndex++

				if byteIndex >= msgLen+startOffset {
					return msg[startOffset : msgLen+startOffset]
				}

				msg = append(msg, 0)
			}
		}
	}
	return
}

func getMessageLength(img *image.NRGBA) (size uint32) {
	var byteSize = decodeNRGBA(0, 4, img)
	size = uint32(byteSize[0])
	size = size << 8
	size = size | uint32(byteSize[1])
	size = size << 8
	size = size | uint32(byteSize[2])
	size = size << 8
	size = size | uint32(byteSize[3])
	return
}

func toNRGBA(r io.Reader) (out *image.NRGBA, err error) {
	var img image.Image

	if img, _, err = image.Decode(r); err != nil {
		return
	}

	out = image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	return
}

func setHeader(msg *[]byte, encodeSize uint32) {
	var mask uint32 = 255
	header := []byte{
		byte(encodeSize >> 24),
		byte((encodeSize >> 16) & mask),
		byte((encodeSize >> 8) & mask),
		byte(encodeSize & mask),
	}
	*msg = append(header, *msg...)
}

func getEachBitMessage(msg []byte, ch chan byte) {
	var (
		byteOffset  int
		bitOffset   int
		currentByte byte
	)

	for {
		if byteOffset >= len(msg) {
			close(ch)
			return
		}

		currentByte = msg[byteOffset]
		ch <- getBitFromByte(currentByte, bitOffset)

		bitOffset++

		if bitOffset >= 8 {
			bitOffset = 0
			byteOffset++
		}
	}
}

func getBitFromByte(b byte, byteIndex int) byte {
	b = b << uint(byteIndex)
	var mask byte = 0x80
	var bit = mask & b

	if bit == 128 {
		return 1
	}
	return 0
}

func setBitInByte(b byte, indexInByte uint32, bit byte) byte {
	var mask byte = 0x80
	mask = mask >> uint(indexInByte)

	if bit == 0 {
		mask = ^mask
		b = b & mask
	} else if bit == 1 {
		b = b | mask
	}
	return b
}

func getLSB(b byte) byte {
	if b%2 == 0 {
		return 0
	}
	return 1
}

func setLSB(b *byte, bit byte) {
	if bit == 1 {
		*b = *b | 1
	} else if bit == 0 {
		var mask byte = 0xFE
		*b = *b & mask
	}
}
