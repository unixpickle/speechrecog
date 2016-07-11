package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/unixpickle/speechrecog/mfcc"
	"github.com/unixpickle/wav"
)

const (
	OutputPerms = 0644
	GraphHeight = 200
	GraphWidth  = 600
)

var CepstrumColors = []string{
	"#8c2323", "#000000", "#d9986c", "#e59900", "#add900", "#4a592d",
	"#00e699", "#2d98b3", "#0061f2", "#cc00ff", "#cc66b8", "#ff0066",
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: mfcc-graph <sound.wav> <output.html> [--velocity]")
		os.Exit(1)
	}

	var getVelocity bool
	if len(os.Args) == 4 {
		if os.Args[3] != "--velocity" {
			fmt.Fprintln(os.Stderr, "Unexpected argument:", os.Args[3])
			os.Exit(1)
		}
		getVelocity = true
	}

	coeffs, err := readCoeffs(os.Args[1], getVelocity)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read MFCCs:", err)
		os.Exit(1)
	}

	page := createHTML(createSVG(coeffs))

	if err := ioutil.WriteFile(os.Args[2], page, OutputPerms); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write result:", err)
		os.Exit(1)
	}
}

func readCoeffs(file string, velocity bool) ([][]float64, error) {
	sound, err := wav.ReadSoundFile(os.Args[1])
	if err != nil {
		return nil, err
	}

	var audioData []float64
	for i, x := range sound.Samples() {
		if i%sound.Channels() == 0 {
			audioData = append(audioData, float64(x))
		}
	}

	mfccSource := mfcc.MFCC(&mfcc.SliceSource{Slice: audioData}, sound.SampleRate(),
		&mfcc.Options{Window: time.Millisecond * 20, Overlap: time.Millisecond * 10})
	if velocity {
		mfccSource = mfcc.AddVelocities(mfccSource)
	}

	var coeffs [][]float64
	for {
		c, err := mfccSource.NextCoeffs()
		if err == nil {
			if velocity {
				coeffs = append(coeffs, c[len(c)/2:])
			} else {
				coeffs = append(coeffs, c)
			}
		} else {
			break
		}
	}

	return coeffs, nil
}

func createSVG(coeffs [][]float64) []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="utf-8" ?>` + "\n")
	buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg"` + "\n")
	buf.WriteString(`     version="1.1"` + "\n")
	buf.WriteString(`     viewBox="0 0 ` + strconv.Itoa(GraphWidth) + " " +
		strconv.Itoa(GraphHeight) + `">` + "\n")

	timeWidth := GraphWidth / float64(len(coeffs))
	var maxVal, minVal float64
	for i, x := range coeffs {
		for j, v := range x[1:] {
			if (i == 0 && j == 0) || v < minVal {
				minVal = v
			}
			if (i == 0 && j == 0) || v > maxVal {
				maxVal = v
			}
		}
	}

	for coeffIdx := 1; coeffIdx < 13; coeffIdx++ {
		color := CepstrumColors[coeffIdx-1]
		buf.WriteString(`  <path id="coeff` + strconv.Itoa(coeffIdx) + `" fill="none" `)
		buf.WriteString(`stroke="` + color + `" d="M`)
		for sampleIdx, timeStep := range coeffs {
			value := timeStep[coeffIdx]
			x := timeWidth * float64(sampleIdx)
			y := GraphHeight * (maxVal - value) / (maxVal - minVal)
			buf.WriteString(formatFloat(x))
			buf.WriteRune(',')
			buf.WriteString(formatFloat(y))
			if sampleIdx+1 < len(coeffs) {
				buf.WriteRune(' ')
			}
		}
		buf.WriteString(`" />` + "\n")
	}

	buf.WriteString("</svg>")
	return buf.Bytes()
}

func createHTML(svgPage []byte) []byte {
	var buf bytes.Buffer

	buf.WriteString("<!doctype html>\n<html>\n")
	buf.WriteString("<head>\n<title>MFCC Graph</title>\n</head>\n<body>\n")
	for coeffIdx := 1; coeffIdx < 13; coeffIdx++ {
		color := CepstrumColors[coeffIdx-1]
		buf.WriteString(`<input type="checkbox" onclick="coeff` + strconv.Itoa(coeffIdx) +
			`.setAttribute('stroke-opacity', checked ? '1' : '0');" checked> ` +
			`<label style="color: ` + color + `">Coeff ` +
			strconv.Itoa(coeffIdx) + "</label>\n")
	}
	buf.Write(svgPage)
	buf.WriteString("\n</body>\n</html>")

	return buf.Bytes()
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
