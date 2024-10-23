package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide the root folder path")
	}

	// Root folder passed as a command-line argument
	rootFolder := os.Args[1]
	listingImagePath(rootFolder)
	generateVideo(rootFolder)
}

func listingImagePath(imagePath string) {
	rootFolder := imagePath

	inputImageFolder := filepath.Join(rootFolder, "assets/images/")
	tempImageList := filepath.Join(rootFolder, "images.txt")

	if err := os.Remove(tempImageList); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	file, err := os.Create(tempImageList)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	err = filepath.Walk(inputImageFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".jpg" {
			absolutePath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			_, err = file.WriteString(fmt.Sprintf("file '%s'\n", absolutePath))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image paths have been written to", tempImageList)
}

func generateVideo(imagePath string) {
	rootFolder := imagePath

	inputAudio := filepath.Join(rootFolder, "assets/audio.mp3")
	inputImageFolder := filepath.Join(rootFolder, "assets/images/")
	//inputImageAssets := filepath.Join(inputImageFolder, "image%d.jpg")
	tempImageList := filepath.Join(rootFolder, "images.txt")
	tempOutputVideo := filepath.Join(rootFolder, "temp_output.mp4")
	outputVideo := filepath.Join(rootFolder, "output.mp4")

	// Step 1: Get the duration of the audio file using ffprobe
	duration, err := getAudioDuration(inputAudio)
	if err != nil {
		log.Fatalf("Failed to get audio duration: %v", err)
	}
	fmt.Println("duration", duration)

	// Step 2: Count total number of images
	totalImages, err := getTotalImages(inputImageFolder)
	if err != nil {
		log.Fatalf("Failed to count images: %v", err)
	}
	fmt.Println("total images:", totalImages)

	// Step 3: Calculate image length (duration per image)
	imageLength := duration / totalImages
	fmt.Println("duration per image:", imageLength)

	// Step 4: Create image video
	err = createImageVideo(tempImageList, imageLength, tempOutputVideo)
	if err != nil {
		log.Fatalf("Failed to create imgage video: %v", err)
	}

	// Step 5: Combine image video with audio
	err = combineVideoWithAudio(tempOutputVideo, inputAudio, outputVideo)
	if err != nil {
		log.Fatalf("Failed to combine video with audio: %v", err)
	}

	// Step 6: remove temporary output video
	err = os.Remove(tempOutputVideo)
	if err != nil {
		log.Fatalf("Failed to remove temp output video: %v", err)
	}

	fmt.Println("generated file:", outputVideo)
}

// Get the duration of the audio using ffprobe
func getAudioDuration(inputAudio string) (int, error) {
	cmd := exec.Command("ffprobe", "-i", inputAudio, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}
	return int(durationFloat), nil
}

// Count the total number of images in the input image folder
func getTotalImages(imageFolder string) (int, error) {
	files, err := os.ReadDir(imageFolder)
	if err != nil {
		return 0, err
	}
	totalImages := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".jpg") {
			totalImages++
		}
	}
	return totalImages, nil
}

// Create image video using ffmpeg
func createImageVideo(tempImagesList string, imageLength int, tempOutputVideo string) error {
	cmd := exec.Command("ffmpeg", "-y", "-safe", "0", "-r", fmt.Sprintf("1/%d", imageLength), "-f", "concat", "-i", tempImagesList, "-c:v", "libx264", "-r", "30", tempOutputVideo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Combine image video with audio using ffmpeg
func combineVideoWithAudio(tempOutputVideo, inputAudio, outputVideo string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", tempOutputVideo, "-i", inputAudio, outputVideo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
