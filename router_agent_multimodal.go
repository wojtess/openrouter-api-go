package openrouterapigo

import (
	"fmt"
	"image"
)

func buildImageContent(messageString string, imgs []image.Image) ([]ContentPart, error) {
	contentList := make([]ContentPart, 0, len(imgs)+1)
	contentList = append(contentList, ContentPart{
		Type: ContentTypeText,
		Text: messageString,
	})
	for _, img := range imgs {
		encodedImage, err := encodeImageToBase64(img)
		if err != nil {
			return nil, err
		}
		contentList = append(contentList, ContentPart{
			Type: ContentTypeImage,
			ImageURL: &ImageURL{
				URL: fmt.Sprintf("data:image/png;base64,%s", encodedImage),
			},
		})
	}
	return contentList, nil
}

func buildPDFContent(messageString string, pathsToPdf []string) ([]ContentPart, error) {
	contentList := make([]ContentPart, 0, len(pathsToPdf)+1)
	contentList = append(contentList, ContentPart{
		Type: ContentTypeText,
		Text: messageString,
	})
	for _, pdfPath := range pathsToPdf {
		encodedPdf, err := encodePDFToBase64(pdfPath)
		if err != nil {
			return nil, err
		}
		contentList = append(contentList, ContentPart{
			Type: ContentTypePDF,
			File: &FileURL{
				Filename: pdfPath,
				FileData: fmt.Sprintf("data:application/pdf;base64,%s", encodedPdf),
			},
		})
	}
	return contentList, nil
}
