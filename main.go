package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/png"
	_ "image/color"
	"image/png"
	"log"
	"os"
	"strings"
	"gonion_ring/model"
)

func main() {
	if parseArguments() {
		process()
	} else {
		printUsage()
	}
}

var sourceName string
var outputName string

func parseArguments() bool {
	var fnCount = 0;
	args := os.Args[1:]
	for _, arg := range args {
		if fnCount == 0 {
			sourceName = arg
			fnCount++
		} else if fnCount == 1 {
			outputName = arg
			fnCount++
		}
	}
	return fnCount == 2
}

func printUsage() {
	fmt.Println("usage: gonion_ring source.png output.png")
}

func process() {
	file, err := os.Open(sourceName)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	source, _, err := image.Decode(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	hash_col := make([]string, source.Bounds().Size().X)
	for x := 0; x < source.Bounds().Size().X; x++ {
		bytes := make([]byte, source.Bounds().Size().Y * 4)
		for y := 0; y < source.Bounds().Size().Y; y++ {
			pixel := source.At(x, y)
			r, g, b, a := pixel.RGBA()
			if a == 0 {
				r = 0
				g = 0
				b = 0
			}
			offset := y * 4
			bytes[offset + 0] = byte(r & 0xFF)
			bytes[offset + 1] = byte(g & 0xFF)
			bytes[offset + 2] = byte(b & 0xFF)
			bytes[offset + 3] = byte(a & 0xFF)
		}
		hasher := sha1.New()
		hasher.Write(bytes)
		hash_col[x] = hex.EncodeToString(hasher.Sum(nil))
	}

	hash_row := make([]string, source.Bounds().Size().Y)
	for y := 0; y < source.Bounds().Size().Y; y++ {
		bytes := make([]byte, source.Bounds().Size().X * 4)
		for x := 0; x < source.Bounds().Size().X; x++ {
			pixel := source.At(x, y)
			r, g, b, a := pixel.RGBA()
			if a == 0 {
				r = 0
				g = 0
				b = 0
			}
			offset := x * 4
			bytes[offset + 0] = byte(r & 0xFF)
			bytes[offset + 1] = byte(g & 0xFF)
			bytes[offset + 2] = byte(b & 0xFF)
			bytes[offset + 3] = byte(a & 0xFF)
		}
		hasher := sha1.New()
		hasher.Write(bytes)
		hash_row[y] = hex.EncodeToString(hasher.Sum(nil))
	}

	range_width := calc_trim_range(hash_col)
	range_height := calc_trim_range(hash_row)

	dpix := 2
	if range_width[1] - range_width[0] <= dpix * 2 {
		range_width = []int{0, -1}
	} else {
		range_width[0] += dpix
		range_width[1] -= dpix
	}
	if range_height[1] - range_height[0] <= dpix * 2 {
		range_height = []int{0, -1}
	} else {
		range_height[0] += dpix
		range_height[1] -= dpix
	}

	create_sliced_image(source, outputName, range_width, range_height)

	var report model.Report

	report.OriginalWidth = source.Bounds().Size().X
	report.OriginalHeight = source.Bounds().Size().Y

	if range_width[0] == 0 && range_width[1] == -1 {
		range_width = []int{1, report.OriginalWidth - dpix}
	}
	if range_height[0] == 0 && range_height[1] == -1 {
		range_height = []int{1, report.OriginalHeight - dpix}
	}
	report.Left = range_width[0] - 1
	report.Top = range_height[0] - 1
	report.Right = report.OriginalWidth - range_width[1] - dpix
	report.Bottom = report.OriginalHeight - range_height[1] - dpix

	// left, top, right, bottom
	s := fmt.Sprintf("%#v", report)
	fmt.Printf("%s\n", strings.Replace(s, "model.Report", "", 1));
}

func calc_trim_range(hash_list []string) []int {
	tmp_hash := ""
	tmp_start_index := 0
	max_length := 0
	max_range := []int{0, -1}
	for index, hash := range hash_list {
		length := ((index - 1) - tmp_start_index)
		if length > max_length {
			max_length = length
			max_range[0] = tmp_start_index
			max_range[1] = index - 1
		}
		if strings.Compare(tmp_hash, hash) != 0 {
			tmp_hash = hash
			tmp_start_index = index
		}
	}
	return max_range
}

func create_sliced_image(source image.Image, output_file_name string, range_width []int, range_height []int) {
	output_width := source.Bounds().Size().X - ((range_width[1] - range_width[0]) + 1)
	output_height := source.Bounds().Size().Y - ((range_height[1] - range_height[0]) + 1)

	rect := image.Rect(0, 0, output_width, output_height)
	output := image.NewNRGBA(rect)
	for ay := 0; ay < output_height; ay++ {
		for ax := 0; ax < output_width; ax++ {
			bx := ax
			by := ay
			if bx >= range_width[0] {
				bx = ax + ((range_width[1] - range_width[0]) + 1)
			}
			if by >= range_height[0] {
				by = ay + ((range_height[1] - range_height[0]) + 1)
			}
			output.Set(ax, ay, source.At(bx, by))
		}
	}

	outFile, openError := os.Create(outputName)
	if openError != nil {
		println("Error", openError)
		return
	}
	writeError := png.Encode(outFile, output)
	if writeError != nil {
		log.Fatal(writeError)
	}
}