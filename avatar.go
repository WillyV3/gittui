package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

// FetchAvatarImage fetches a GitHub avatar and returns the resized image
func FetchAvatarImage(avatarURL string, size int) (image.Image, error) {
	// Fetch the image
	resp, err := http.Get(avatarURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch avatar: status %d", resp.StatusCode)
	}

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if necessary (GitHub avatars are typically 460x460, we want smaller for terminal)
	img = resizeImage(img, size)

	return img, nil
}

// resizeImage resizes an image to fit within maxSize while maintaining aspect ratio
func resizeImage(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is already small enough, return as-is
	if width <= maxSize && height <= maxSize {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = (height * maxSize) / width
	} else {
		newHeight = maxSize
		newWidth = (width * maxSize) / height
	}

	// Create a new image with the target size
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor sampling for resizing
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Map to source coordinates
			srcX := (x * width) / newWidth
			srcY := (y * height) / newHeight
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	return newImg
}
