package utils

import (
	"bytes"
	"demo-go/config"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-resty/resty/v2"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func IsUrl(inputPath string) bool {
	return strings.HasPrefix(inputPath, "http://") || strings.HasPrefix(inputPath, "https://")
}

func ReadImage(inputPath string) (image.Image, error) {
	/** 判断文件是本地路径还是url&&读取文件 */
	var err error
	var fileOrigin interface {
		Close() error
		Read([]byte) (int, error)
	}

	if IsUrl(inputPath) {
		cli := resty.New()
		resp, err := cli.R().Get(inputPath)
		if err != nil {
			fmt.Println("GET(file)错误")
			log.Fatal(err)
		}
		fileOrigin = resp.RawResponse.Body

	} else {
		fileOrigin, err = os.Open(inputPath)
	}

	defer fileOrigin.Close()
	if err != nil {
		fmt.Println("os.Open(file)错误")
		log.Fatal(err)
	}
	var origin image.Image

	ext := filepath.Ext(inputPath)
	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") {
		/** jpg 格式 */
		origin, err = jpeg.Decode(fileOrigin)
		if err != nil {
			fmt.Println("jpeg.Decode(file_origin)")
			log.Fatal(err)
		}
	} else if strings.EqualFold(ext, ".png") {
		origin, err = png.Decode(fileOrigin)
		if err != nil {
			fmt.Println("png.Decode(file_origin)")
			log.Fatal(err)
		}
	} else {
		return nil, fmt.Errorf("输入图片类型错误")
	}
	return origin, err
}

func ResizeImage(img image.Image, scale float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	imageResize := resize.Resize(uint(Multiply(width, scale)), uint(Multiply(height, scale)), img, resize.Lanczos3)
	return imageResize
}

// type MyBuffer struct {
// 	buffer *bytes.Buffer
// }

// func (b MyBuffer) Write(p []byte) (int, error) {
// 	return b.buffer.Write(p)
// }
// func (b MyBuffer) Read(p []byte) (int, error) {
// 	return b.buffer.Read(p)
// }
// func (b MyBuffer) Close() error {
// 	return nil
// }

func SaveImage(outPath string, src image.Image) error {
	var err error
	ext := filepath.Ext(outPath)
	if IsUrl(outPath) {
		buffer := new(bytes.Buffer)
		outPath = strings.Split(outPath, "/hsr/")[1]

		err = encodeImage(ext, buffer, src)
		if err != nil {
			return err
		}
		//err = metaloop.SaveImageToS3("image/"+ext, outPath, int64(buffer.Len()), buffer)

	} else {
		f, err := os.OpenFile(outPath, os.O_SYNC|os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()

		err = encodeImage(ext, f, src)
	}

	return err
}

func SaveNewImage(imagePath string, image *image.RGBA, quality int) error {
	outFile, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	err = jpeg.Encode(outFile, image, &jpeg.Options{Quality: quality})
	if err != nil {
		return err
	}
	return nil
}

func encodeImage(ext string, f io.Writer, src image.Image) error {
	var err error
	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") {
		err = jpeg.Encode(f, src, &jpeg.Options{Quality: 60})
		if err != nil {
			fmt.Println("jpeg encode failed")
		}
	} else if strings.EqualFold(ext, ".png") {
		// create buffer
		buff := new(bytes.Buffer)

		// encode image to buffer
		err = png.Encode(buff, src)
		if err != nil {
			fmt.Println("failed to create buffer", err)
		}
		imgSrc, _, err := image.Decode(bytes.NewReader(buff.Bytes()))
		if err != nil {
			fmt.Println("image to []bytes failed")
		}

		//convert png to jpg
		newImg := image.NewRGBA(imgSrc.Bounds())
		draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
		draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)

		err = jpeg.Encode(f, newImg, &jpeg.Options{Quality: 40})
		if err != nil {
			fmt.Println("jpeg encode failed")
		}
	}
	return err
}

func CompressImage(inputPath string, cover bool, method int) (string, error) {
	var err error
	// "abc/def.aaa/bbb/c.jpg"
	fileName := path.Base(inputPath)
	filePath := path.Dir(inputPath)
	fileNameSplit := strings.Split(fileName, ".")
	if len(fileNameSplit) < 2 {
		fmt.Printf("输入图片路径不合法，inputPath %v\n", inputPath)
		return inputPath, err
	}
	outPath := path.Join(filePath, fileNameSplit[0]+"_compress.jpg")
	if Exists(outPath) && cover == false {
		return outPath, nil
	}
	if method == 0 {
		origin, err := ReadImage(inputPath)
		if err != nil {
			fmt.Printf("读取源图片失败，inputPath %v\n", inputPath)
			return inputPath, err
		}
		err = SaveImage(outPath, origin)
		if err != nil {
			return inputPath, err
		}
	} else if method == 1 {
		alConf := config.Conf.Algorithm
		type compressRequest struct {
			ImagePath string `json:"image_path"`
		}
		body := compressRequest{ImagePath: inputPath}
		var response struct{ Status string }
		_, err := HTTPClient("POST", fmt.Sprintf("http://%s:%d/api/compress", alConf.IP, alConf.Port), body, response)
		if err != nil {
			return inputPath, err
		}
	}
	return outPath, err
}

func DepthRendering(depthURL string, imageURL string, cover bool) (string, error) {
	// DepthRender 深度图渲染
	type DepthRender struct {
		DepthURL string `json:"depth_url"`
		ImageURL string `json:"image_url"`
		Output   string `json:"output"`
	}

	type HTTPResponse struct {
		Status  string `json:"status"`
		Code    int    `json:"code"`
		Message string `json:"message"`
		OutPath string `json:"out_path"`
	}

	var outPath, renderURL string
	renderURL = strings.ReplaceAll(depthURL, "D.tif", "rendered.jpg")
	if strings.HasSuffix(depthURL, "D.exr") {
		renderURL = strings.ReplaceAll(depthURL, "D.exr", "rendered.jpg")
	}

	// if strings.Contains(depthURL, "/hsr/") {
	// 	outPath =strings.Replace(depthURL, "D.tif" , "rendered.jpg")
	// } else {
	// 	depthPath := strings.Split(depthURL, "files/")[1]
	// 	outPath = "files/" + strings.Split(depthPath, "D.tif")[0] + "rendered.jpg"
	// 	renderURL = fmt.Sprintf("http://%s:%d/%s", configs.Conf.APP.IP, configs.Conf.APP.Port, outPath)
	// }

	if !cover && Exists(renderURL) {

		// renderURL := fmt.Sprintf("http://%s:%d/%s", configs.Conf.APP.IP, configs.Conf.APP.Port, outPath)
		return renderURL, nil
	}
	if !strings.Contains(depthURL, "/hsr/") {
		depthPath := strings.Split(depthURL, "files/")[1]
		outPath = "files/" + strings.Split(depthPath, "D.tif")[0] + "rendered.jpg"
		if strings.HasSuffix(depthURL, "D.exr") {
			outPath = strings.ReplaceAll(depthPath, "D.exr", "rendered.jpg")
		}

		outPathSplit := strings.Split(outPath, "/")
		outDir := strings.Join(outPathSplit[:len(outPathSplit)-1], "/")
		if !IsExist(outDir) {
			err := os.MkdirAll(outDir, os.ModePerm)
			if err != nil {
				return "", err
			}
		}
	}

	url := fmt.Sprintf("http://%s:%d/api/depth_to_render", config.Conf.Algorithm.IP, config.Conf.Algorithm.Port)
	requestData := DepthRender{DepthURL: depthURL, ImageURL: imageURL, Output: renderURL}
	renderedDepthResponse := HTTPResponse{}
	code, err := HTTPClient("POST", url, &requestData, &renderedDepthResponse)

	if err != nil || code != 200 {
		return imageURL, nil
	} else {
		// renderURL := fmt.Sprintf("http://%s:%d/%s", configs.Conf.APP.IP, configs.Conf.APP.Port, outPath)
		return renderURL, nil
	}
}

// CompressJPEGImage 可用于同时生成多张质量不同的压缩图
func CompressJPEGImage(imagePath string, quality []int, cover bool) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 解码图片文件
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	for _, qua := range quality {
		outputPath := strings.Replace(imagePath, ".jpg", fmt.Sprintf("_compressed%d.jpg", qua), 1)
		if cover == false && Exists(outputPath) {
			continue
		}
		// 创建输出文件
		out, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}

		// 重新编码图片并指定输出质量
		err = imaging.Encode(out, img, imaging.JPEG, imaging.JPEGQuality(qua))
		if err != nil {
			return err
		}
		err = out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
