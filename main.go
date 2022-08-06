package main

import (
	"encoding/binary"
	"log"
	"os"
)

type pixel struct {
	r byte
	b byte
	g byte
}

func main() {
	inputFileName := os.Args[1]

	inputFile, err := os.ReadFile(inputFileName)
	if err != nil {
		log.Fatal(err)
	}

	fileHeader := inputFile[0:14]
	infoHeader := inputFile[14:54]
	bmp := inputFile[54:]
	wB := binary.LittleEndian.Uint16(infoHeader[4:8])
	hB := binary.LittleEndian.Uint16(infoHeader[8:12])
	hPx := int32(int16(hB) * -1)
	wPx := int16(wB)
	bytesWide := int32(wPx * 3)
	total := hPx * bytesWide

	outputFile, err := os.OpenFile("filtered.bmp", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// writes the header to the outputFile
	_, headerErr := outputFile.Write(fileHeader)
	if err != nil {
		log.Fatal("failed to write header", headerErr)
	}
	_, infoHError := outputFile.Write(infoHeader)
	if err != nil {
		log.Fatal("failed to write header", infoHError)
	}

	// performs the filter op based on the arg
	filteredBMP := []byte{}
	// convert to grayscale
	if os.Args[2] == "-g" {
		for i := 0; i < len(bmp)-2; i += 3 {
			pxAvg := bmp[i]/3 + bmp[i+1]/3 + bmp[i+2]/3
			filteredBMP = append(filteredBMP, pxAvg)
			filteredBMP = append(filteredBMP, pxAvg)
			filteredBMP = append(filteredBMP, pxAvg)
		}
	}

	// mirror image
	if os.Args[2] == "-r" {
		for i := int32(0); i < total; i += bytesWide {
			row := []byte{}
			for j := int32(0); j < bytesWide-2; j += 3 {
				row = append([]byte{bmp[i+j], bmp[i+(j+1)], bmp[i+(j+2)]}, row...)
			}
			filteredBMP = append(filteredBMP, row...)
		}
	}

	// box blur
	if os.Args[2] == "-b" {
		pxArr := []pixel{}
		for i := 0; i < len(bmp)-2; i += 3 {
			pxArr = append(pxArr, pixel{bmp[i], bmp[i+2], bmp[i+1]})
		}

		width := int(wB)

		blurred := boxBlur(pxArr, width)

		for _, pix := range blurred {
			filteredBMP = append(filteredBMP, pix.r, pix.g, pix.b)
		}

		_, wErr := outputFile.Write(filteredBMP)
		if wErr != nil {
			log.Fatal("something went wrong")
		}
	}
}

func boxBlur(img []pixel, w int) []pixel {
	ret := []pixel{}

	for i := range img {
		kernel := calcKernel(img, i, w)
		avg := calcAvg(kernel)
		ret = append(ret, avg)
	}

	return ret
}

// func sobel(img []pixel, w int) []pixel {
// 	ret := []pixel{}

// 	for i := range img {
// 		kernel := calcKernel(img, i, w)
// 		sobel := calcSobel(kernel)
// 	}
// }

func calcKernel(img []pixel, i int, w int) []pixel {
	// // test
	// if i%w != 0 && (i+1)%w != 0 && len(img)-i >= w && i > w { // return for all non-edges
	// 	return []pixel{
	// 		img[i-w-1], //NW
	// 		img[i-w],   //N
	// 		img[i-w+1], //NE
	// 		img[i-1],   //W
	// 		img[i],     //CURR
	// 		img[i+1],   //E
	// 		img[i+w-1], //SW
	// 		img[i+w],   //S
	// 		img[i+w+1], //SE
	// 		// return []pixel{}
	// 	}

	// } else {
	// 	return []pixel{}
	// }

	if i < w && i%w == 0 { // top left corner
		return []pixel{
			img[i],     //NW but use CURR
			img[i],     //N but use CURR
			img[i+1],   //NE but use E
			img[i],     //W but use CURR
			img[i],     //CURR
			img[i+1],   //E
			img[i+w],   //SW but use S
			img[i+w],   //S
			img[i+w+1], //SE
		}
	} else if i < w && (i+1)%w == 0 { // top right corner
		return []pixel{
			img[i-1],   //NW but use W
			img[i],     //N but use CURR
			img[i],     //NE but use CURR
			img[i-1],   //W
			img[i],     //CURR
			img[i],     //E but use CURR
			img[i+w-1], //SW
			img[i+w],   //S
			img[i+w],   //SE but use S
		}
	} else if i > len(img)-w && i%w == 0 { // bottom left corner
		return []pixel{
			img[i-w],   //NW but use N
			img[i-w],   //N
			img[i-w+1], //NE
			img[i],     //W but use CURR
			img[i],     //CURR
			img[i+1],   //E
			img[i],     //SW but use CURR
			img[i],     //S but use CURR
			img[i+1],   //SE but use E
		}
	} else if i == len(img)-1 { // bottom right orner
		return []pixel{
			img[i-w-1], //NW
			img[i-w],   //N
			img[i-w],   //NE but use N
			img[i-1],   //W
			img[i],     //CURR
			img[i],     //E but use CURR
			img[i-1],   //SW but use W
			img[i],     //S but use CURR
			img[i],     //SE but use CURR
		}
	} else if i < w { // top border
		return []pixel{
			img[i-1],   //NW but use W
			img[i],     //N but use curr
			img[i+1],   //NE but use E
			img[i-1],   //W
			img[i],     //CURR
			img[i+1],   //E
			img[i+w-1], //SW
			img[i+w],   //S
			img[i+w+1], //SE
		}
	} else if i%w == 0 && len(img)-i > w { //left edge
		return []pixel{
			img[i-w],   //NW but use N
			img[i-w],   //N
			img[i-w+1], //NE
			img[i],     //W but use CURR
			img[i],     //CURR
			img[i+1],   //E
			img[i+w],   //SW but use S
			img[i+w],   //S
			img[i+w+1], //SE
		}
	} else if (i+1)%w == 0 { //right edge
		return []pixel{
			img[i-w-1], //NW
			img[i-w],   //N
			img[i-w],   //NE but use N
			img[i-1],   //W
			img[i],     //CURR
			img[i],     //E but use CURR
			img[i+w-1], //SW
			img[i+w],   //S
			img[i+w],   //SE but use S
		}
	} else if len(img)-i <= w { //bottom border
		return []pixel{
			img[i-w-1], //NW
			img[i-w],   //N
			img[i-w+1], //NE
			img[i-1],   //W
			img[i],     //CURR
			img[i+1],   //E
			img[i-1],   //SW but use W
			img[i],     //S but use CURR
			img[i+1],   //SE but use E
		}
	} else if i%w != 0 && (i+1)%w != 0 && len(img)-i >= w && i > w { // return for all non-edges
		return []pixel{
			img[i-w-1], //NW
			img[i-w],   //N
			img[i-w+1], //NE
			img[i-1],   //W
			img[i],     //CURR
			img[i+1],   //E
			img[i+w-1], //SW
			img[i+w],   //S
			img[i+w+1], //SE
			// return []pixel{}
		}
	} else {
		return []pixel{}
	}
}

func calcAvg(kernel []pixel) pixel {
	r := byte(0)
	g := byte(0)
	b := byte(0)

	for _, pix := range kernel {
		r += pix.r / 9
		g += pix.g / 9
		b += pix.b / 9
	}

	return pixel{
		r: r,
		g: g,
		b: b,
	}
}
