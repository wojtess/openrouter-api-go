package openrouterapigo

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
)

func encodeImageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer

	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
}

func encodePDFToBase64(pdfPath string) (string, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}
