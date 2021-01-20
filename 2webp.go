package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/cjun714/go-image/tga"
	"github.com/cjun714/go-image/webp"
)

var quality = 85

func main() {
	src := os.Args[1]
	targetDir := os.Args[1]

	if len(os.Args) == 3 {
		targetDir = os.Args[2]
	}

	var wg sync.WaitGroup
	e := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if !isImage(path) {
			return nil
		}

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			f, e := os.Open(path)
			if e != nil {
				fmt.Printf("open %s failed, error:%s\n", path, e)
				return
			}
			defer f.Close()
			img, imgType, e := image.Decode(f)
			if e != nil {
				fmt.Printf("decode image failed, error:%s\n", e)
				return
			}

			if isNormal(img) {
				fmt.Println("normal map:", path)

				if imgType == "jpeg" { // skip .jpg normal map
					return
				}
				quality = 100
			}

			ext := filepath.Ext(path)
			name := strings.TrimSuffix(path, ext)
			name = name + ".webp"
			name = filepath.Join(targetDir, filepath.Base(name))

			wr, e := os.Create(name)
			if e != nil {
				fmt.Printf("create file %s failed, error:%s\n", name, e)
				return
			}
			defer wr.Close()
			cfg, e := webp.ConfigPreset(webp.PRESET_PHOTO, quality)
			if e != nil {
				fmt.Printf("crate webp config failed, %s, error:%s\n", name, e)
				return
			}
			if e = webp.Encode(wr, img, cfg); e != nil {
				fmt.Printf("encode %s failed, error:%s\n", name, e)
			}
		}(path)

		return nil
	})
	wg.Wait()

	if e != nil {
		panic(e)
	}
}

var imgExt = []string{
	".jpg",
	".jpeg",
	".png",
	".tga",
	".bmp",
}

func isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, iext := range imgExt {
		if ext == iext {
			return true
		}
	}

	return false
}

func isNormal(img image.Image) bool {
	var r, g, b uint8

	switch t := img.(type) {
	case *image.NRGBA:
		r, g, b = checkRGBA(t.Pix)
	case *image.RGBA:
		r, g, b = checkRGBA(t.Pix)
	case *image.YCbCr:
		r, g, b = checkYCbCr(t)
	default:
		return false
	}

	dr, dg, db := int32(r)-127, int32(g)-127, 255-int32(b)
	if (-30 < dr && dr < 30) && (-30 < dg && dg < 30) && db < 60 {
		return true
	}

	return false
}

func checkRGBA(bts []byte) (uint8, uint8, uint8) {
	r, g, b := 0.0, 0.0, 0.0

	length := len(bts) / 4
	count := 0

	for i := 0; i < length; i++ {
		if bts[i*4+3] == 0 { // if alpha < 10, skip
			continue
		}
		r += float64(bts[i*4])
		g += float64(bts[i*4+1])
		b += float64(bts[i*4+2])
		count++
	}
	r, g, b = r/float64(count), g/float64(count), b/float64(count)

	return uint8(r), uint8(g), uint8(b)
}

func checkYCbCr(img *image.YCbCr) (uint8, uint8, uint8) {
	w := img.Bounds().Size().X
	h := img.Bounds().Size().Y

	r, g, b := 0.0, 0.0, 0.0
	count := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.YCbCrAt(x, y)
			R, G, B := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
			if R == 0 && G == 0 && B == 0 || R == 255 && G == 255 && B == 255 {
				continue
			}
			r += float64(R)
			g += float64(G)
			b += float64(B)

			count++
		}
	}

	r, g, b = r/float64(count), g/float64(count), b/float64(count)

	return uint8(r), uint8(g), uint8(b)
}
