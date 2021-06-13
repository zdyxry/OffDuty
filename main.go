package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"path/filepath"

	_ "image/jpeg"
	"image/png"
)

func convertPhoto(filePath string) (string, error) {

	extension := filepath.Ext(filePath)
	targetFilePath := filePath[0:len(filePath)-len(extension)] + "_gray.png"
	log.Printf("Target File Path is: %s", targetFilePath)

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	log.Printf("Image type: %T", img)

	// Converting image to grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	// Working with grayscale image, e.g. convert to png
	f, err = os.Create(targetFilePath)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, grayImg); err != nil {
		log.Fatal(err)
	}

	return targetFilePath, nil
}

func updateSlackProfile(token string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return err
	}
	_ = writer.WriteField("token", token)

	var slackURL = "https://slack.com/api/users.setPhoto"
	req, err := http.NewRequest("POST", slackURL, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("resp err: ", err)
		return err
	}
	return nil
}

func main() {
	grayPtr := flag.Bool("gray", false, "Photo with gray")
	slackTokenPtr := flag.String("slackToken", "", "Token for Slack(App)")
	photoPathPtr := flag.String("photoPath", "", "Profile photo path")

	flag.Parse()
	fmt.Println("Flag gray:", *grayPtr)
	fmt.Println("Flag slackToken:", *slackTokenPtr)
	fmt.Println("Flag photoPath:", *photoPathPtr)

	targetFilePath, err := filepath.Abs(*photoPathPtr)
	if err != nil {
		log.Fatal(err)
	}
	if *grayPtr == true {
		targetFilePath, err = convertPhoto(targetFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Profile photo path is: %s", targetFilePath)

	if *slackTokenPtr != "" {
		err := updateSlackProfile(*slackTokenPtr, targetFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

}
