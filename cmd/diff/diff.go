package diff

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"strings"

	diffimage "github.com/schollz/go-diff-image"
)

const ssDir = "screenshots"

func decodePNG(png string) (image.Image, error) {
	f, err := os.Open(png)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode image: %v", err)
	}
	return img, nil
}

func compareImage(fromImagePath string, toImagePath string, diffDir string) error {
	fromImg, err := decodePNG(fromImagePath)
	if err != nil {
		return fmt.Errorf("could not decode image %s : %v", fromImagePath, err)
	}
	toImg, err := decodePNG(toImagePath)
	if err != nil {
		return fmt.Errorf("could not decode image %s : %v", toImagePath, err)
	}

	dst, _, _, _ := diffimage.DiffImage(fromImg, toImg)

	_, fromPng := path.Split(fromImagePath)

	diffImageName := "diff-" + strings.Replace(fromPng, ".png", "", -1) + ".png"
	diffImagePath := path.Join(diffDir, diffImageName)
	f, err := os.OpenFile(diffImagePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open diff file: %v", err)
	}
	defer f.Close()
	png.Encode(f, dst)
	return nil
}

func Diff(from string, to string) error {
	fromDir := path.Join(ssDir, from)
	if _, err := os.Stat(fromDir); os.IsNotExist(err) {
		return err
	}
	toDir := path.Join(ssDir, to)
	if _, err := os.Stat(toDir); os.IsNotExist(err) {
		return err
	}
	diffDir := path.Join(ssDir, from+"-"+to)

	files, err := os.ReadDir(fromDir)
	if err != nil {
		return err
	}
	var fileList []string
	for _, f := range files {
		if path.Ext(f.Name()) == ".png" {
			fileList = append(fileList, f.Name())
		}
	}

	for _, f := range fileList {
		if _, err := os.Stat(path.Join(toDir, f)); err == nil {
			err := compareImage(path.Join(fromDir, f), path.Join(path.Join(toDir, f)), diffDir)
			if err != nil {
				fmt.Printf("failed to compare %s and %s: %v", path.Join(fromDir, f), path.Join(toDir, f), err)
			}
		}
	}
	return nil
}
