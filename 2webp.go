package main

import "C"
import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/cjun714/go-image-stb/stb"
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

			pixPtr, w, h, comps, e := stb.Load(path)
			defer stb.Free(pixPtr)
			if e != nil {
				fmt.Printf("encode image failde %s", path)
				return
			}

			pix := C.GoBytes(unsafe.Pointer(pixPtr), C.int(w*h*comps))
			if isNormal(pix, comps) {
				fmt.Println("normal map:", path)

				if strings.HasSuffix(path, ".jpeg") { // skip .jpg normal map
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
			if e = webp.EncodePixBytes(wr, pix, w, h, comps, cfg); e != nil {
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

func isNormal(pix []byte, comps int) bool {
	var r, g, b uint8

	switch comps {
	case 3:
		r, g, b = checkRGB(pix)
	case 4:
		r, g, b = checkRGBA(pix)
	default:
		return false
	}

	dr, dg, db := int32(r)-127, int32(g)-127, 255-int32(b)
	if (-30 < dr && dr < 30) && (-30 < dg && dg < 30) && db < 60 {
		return true
	}

	return false
}

func checkRGB(bts []byte) (uint8, uint8, uint8) {
	r, g, b := 0.0, 0.0, 0.0

	length := len(bts) / 3
	count := 0

	for i := 0; i < length; i++ {
		r += float64(bts[i*3])
		g += float64(bts[i*3+1])
		b += float64(bts[i*3+2])
		count++
	}
	r, g, b = r/float64(count), g/float64(count), b/float64(count)

	return uint8(r), uint8(g), uint8(b)
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
