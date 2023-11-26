package main

import (
	"fmt"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mesilliac/pulse-simple"
	"strconv"
	"time"
)

var audioBuffer []byte
var lastFrame time.Time

var (
	rate            int32 = size * 8
	channels        int32 = 2
	size            int32 = 1024
	horizontal            = true
	colorTransition       = false
	hidden                = false
)

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint)

	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1024, 512, "waveform")

	defer rl.CloseWindow()

	ss := pulse.SampleSpec{
		Format:   pulse.SAMPLE_U8,
		Rate:     uint32(rate),
		Channels: uint8(channels),
	}

	s, err := pulse.Capture("waveform", "waveform", &ss)

	var stream = s

	if err != nil {
		panic(err)
	}

	defer stream.Free()

	audioBuffer = make([]byte, size)

	var (
		oldRate = rate
		oldSize = size
	)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		if rl.IsKeyPressed(rl.KeyF1) {
			hidden = !hidden
		}

		// todo: move this to a goroutine to avoid blocking the main and dropping fps
		updateAudio(&audioBuffer, stream)

		drawFps()

		drawHelpText()

		rl.ClearBackground(rl.Black)

		if horizontal {
			for i := int32(1); i <= channels; i++ {
				for j := int32(1); j <= size/channels; j++ {
					rl.DrawRectangle(int32(j*2), int32(audioBuffer[j]), 5, 2, rl.Color{R: 255, G: 255, B: 255, A: 255})
				}
			}
		} else {
			for i := int32(1); i <= channels; i++ {
				for j := int32(1); j <= size/channels; j++ {
					rl.DrawRectangle(int32(audioBuffer[j])+200, int32(j*2), 2, 5, rl.Color{R: 255, G: 255, B: 255, A: 255})
				}
			}
		}

		rate = updateRateSlider()
		size = updateSizeSlider()

		horizontal = updateHorizontalCheckbox()

		if rate != oldRate {
			size = calculateNewSize(rate, channels)

			ss = pulse.SampleSpec{
				Format:   pulse.SAMPLE_U8,
				Rate:     uint32(rate),
				Channels: uint8(channels),
			}

			stream.Free()

			s, err = pulse.Capture("waveform", "waveform", &ss)
			if err != nil {
				panic(err)
			}

			stream = s

			audioBuffer = make([]byte, size)
		}

		if size != oldSize {
			ss = pulse.SampleSpec{
				Format:   pulse.SAMPLE_U8,
				Rate:     uint32(calculateNewRate(size, channels)),
				Channels: uint8(channels),
			}

			stream.Free()

			s, err = pulse.Capture("waveform", "waveform", &ss)
			if err != nil {
				panic(err)
			}

			stream = s

			audioBuffer = make([]byte, size)
		}

		oldRate = rate
		oldSize = size

		rl.EndDrawing()
	}
}

func drawFps() {
	if hidden {
		return
	}
	fps := fmt.Sprintf("FPS: %d", rl.GetFPS())
	raygui.TextBox(rl.NewRectangle(900, 450, 60, 20), &fps, len(fps), false)
}

func drawHelpText() {
	if hidden {
		return
	}

	raygui.Label(rl.NewRectangle(900, 400, 200, 20), "F1: Toggle UI")
}

func updateHorizontalCheckbox() bool {
	if hidden {
		return horizontal
	}

	return raygui.CheckBox(rl.NewRectangle(75, 250, 20, 20), "Horizontal", horizontal)
}

func updateRateSlider() int32 {
	if hidden {
		return rate
	}

	return int32(raygui.SliderBar(rl.NewRectangle(75, 300, 200, 20), "Rate: "+strconv.Itoa(int(rate)), "", float32(rate), 1024, 192000))
}

func updateSizeSlider() int32 {
	if hidden {
		return size
	}

	return int32(raygui.SliderBar(rl.NewRectangle(75, 350, 200, 20), "Size: "+strconv.Itoa(int(size)), "", float32(size), 1024, 16384))
}

func calculateNewRate(size int32, channels int32) int32 {
	return size * channels * 8
}

func calculateNewSize(rate int32, channels int32) int32 {
	return rate / channels / 8
}

func updateAudio(audioBuffer *[]byte, stream *pulse.Stream) {
	_, err := stream.Read(*audioBuffer)
	if err != nil {
		panic(err)
	}
}
