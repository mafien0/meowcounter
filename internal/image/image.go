// Package image concats counter images
package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sync"

	"meowcounter/assets"
)

var ErrToMuch = errors.New("number is too much! increase numCount")

// Cache!
var (
	mu    sync.Mutex
	cache = map[rune]image.Image{}
)

func load(num rune) (image.Image, error) {
	// Read from cache first
	mu.Lock()
	if img, ok := cache[num]; ok {
		mu.Unlock()
		return img, nil
	}
	mu.Unlock()

	// Read from the embed
	path, ok := m[num]
	if !ok {
		return nil, fmt.Errorf("no file found for %v", num)
	}
	data, err := assets.FS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Cache it
	mu.Lock()
	cache[num] = img
	mu.Unlock()
	return img, nil
}

func Glue(num string, numCount int) (image.Image, error) {
	if len(num) > numCount {
		return nil, ErrToMuch
	}

	var digits []image.Image

	// Fill empty digits with 0
	unused := numCount - len(num)
	for range unused {
		img, err := load('0')
		if err != nil {
			return nil, err
		}
		digits = append(digits, img)
	}

	// Append each digit
	for _, v := range num {
		img, err := load(v)
		if err != nil {
			return nil, err
		}
		digits = append(digits, img)
	}

	// Glue em'

	// Calculate bounds
	w, h := 0, 0
	for _, img := range digits {
		b := img.Bounds()
		w += b.Dx()
		if b.Dy() > h {
			h = b.Dy()
		}
	}

	// Draw
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	x := 0
	for _, img := range digits {
		b := img.Bounds()
		draw.Draw(dst, image.Rect(x, 0, x+b.Dx(), b.Dy()), img, b.Min, draw.Src)
		x += b.Dx()
	}
	return dst, nil
}
