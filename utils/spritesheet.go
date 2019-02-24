/*
Sheet.xml contains data on the open source CC0 sprites I am using for this project 
*/

package utils
// RENAME package to package utils after testing is complete

import (
    "encoding/xml"
    "fmt"
    "io"
    "os"
	"path/filepath"
	"strconv"
)

// Properties of a XMLImage are defined by the spritesheet
// To use ebiten to render images the following properties are required,
// start x, start y, width and height in order to draw a rectangle.
type XMLImage struct {
	XMLName  xml.Name `xml:"SubTexture"`
	Name     string   `xml:"name,attr"`
	X        string   `xml:"x,attr"`
	Y        string   `xml:"y,attr"`
	Width    string   `xml:"width,attr"`
	Height   string   `xml:"height,attr"`
}

type XMLImages struct {
	XMLName   xml.Name    `xml:"TextureAtlas"`
	imagePath string      `xml:"imagePath,attr"`
	Images    []XMLImage  `xml:"SubTexture"`
}

/**
 * how to render the image struct, honestly this is not very helpful
 * simply because the image path must be inputted to read xml.
 *
 */
func ReadImagesStruct(reader io.Reader) (*XMLImages, error) {
    var xmlImages XMLImages
    if err := xml.NewDecoder(reader).Decode(&xmlImages); err != nil {
        return nil, err
    }

    return &xmlImages, nil
}

/**
 * how to render the image struct, honestly this is not very helpful
 * simply because the image path must be inputted to read xml.
 *
 */
 func ReadImages(reader io.Reader) ([]XMLImage, error) {
    var xmlImages XMLImages
    if err := xml.NewDecoder(reader).Decode(&xmlImages); err != nil {
        return nil, err
    }

    return xmlImages.Images, nil
}

// TODO Replace os.exit with error message
func ReadImageData(imagePath string) ([]XMLImage, error) {
    // Build the location of the sheet.xml file
	// filepath.Abs appends the file name to the default working directly,
	// take input parameter to path.

    imagesFilePath, err := filepath.Abs(imagePath)
    if err != nil {
        fmt.Println(err)
        // os.Exit(1)
    }

    // Open the straps.xml file
    file, err := os.Open(imagesFilePath)
    if err != nil {
        fmt.Println(err)
        // os.Exit(1)
    }

    defer file.Close()

    // Read the images file
    xmlImages, err := ReadImages(file)
    if err != nil {
        fmt.Println(err)
        // os.Exit(1)
    }

	return xmlImages, nil

}

// returns subset of images for animation.
func imageSplice(xmlImages []XMLImage,start int ,end int) ([]XMLImage, error) {
	
	testNum := start 
	endNum  := end
	// consider adding error handling for out of bounds case
	return xmlImages[testNum:endNum], nil
}

// unpack struct data from XMLImage
func ImageData(image XMLImage) (string, int, int, int, int) {
	name      := image.Name
	x, _      := strconv.Atoi(image.X)
	y, _      := strconv.Atoi(image.Y)
	width,  _ := strconv.Atoi(image.Width)
	height, _ := strconv.Atoi(image.Height)
	// consider checking if any of the values are null, or maybe go handles that for me.
	return name, x, y, width, height
}

// DELETE LATER, testing function
func main() {
    // Build the location of the straps.xml file
	// filepath.Abs appends the file name to the default working directly
	xmlImages, _  := ReadImageData("../resources/sheet.xml")
	beamImages, _ := imageSplice(xmlImages,0,7)
    for i := 0; i < 7; i++ {
		name, x ,y, _, _ := ImageData(beamImages[i])
		fmt.Printf("Name: %s  x: %d y %d", name, x, y)
	}
}
