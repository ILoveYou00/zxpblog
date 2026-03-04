package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
	"time"
)

// Captcha 验证码结构
type Captcha struct {
	Code  string
	Image string
}

// GenerateCaptcha 生成验证码图片
func GenerateCaptcha(width, height int) *Captcha {
	// 生成随机验证码
	code := generateCode(4)

	// 创建图片
	img := createCaptchaImage(code, width, height)

	// 转换为 base64
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	base64Img := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	return &Captcha{
		Code:  code,
		Image: base64Img,
	}
}

// generateCode 生成随机验证码
func generateCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 去除容易混淆的字符
	code := ""
	for i := 0; i < length; i++ {
		code += string(chars[r.Intn(len(chars))])
	}
	return code
}

// createCaptchaImage 创建验证码图片
func createCaptchaImage(code string, width, height int) *image.RGBA {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建图片
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景色（浅色）
	bgColor := color.RGBA{
		R: uint8(240 + r.Intn(15)),
		G: uint8(240 + r.Intn(15)),
		B: uint8(240 + r.Intn(15)),
		A: 255,
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 添加干扰线
	for i := 0; i < 5; i++ {
		x1 := r.Intn(width)
		y1 := r.Intn(height)
		x2 := r.Intn(width)
		y2 := r.Intn(height)
		lineColor := color.RGBA{
			R: uint8(r.Intn(200)),
			G: uint8(r.Intn(200)),
			B: uint8(r.Intn(200)),
			A: 255,
		}
		drawLine(img, x1, y1, x2, y2, lineColor)
	}

	// 添加干扰点
	for i := 0; i < 100; i++ {
		x := r.Intn(width)
		y := r.Intn(height)
		dotColor := color.RGBA{
			R: uint8(r.Intn(200)),
			G: uint8(r.Intn(200)),
			B: uint8(r.Intn(200)),
			A: 255,
		}
		img.Set(x, y, dotColor)
	}

	// 绘制验证码文字（简化版，使用固定位置）
	charWidth := width / len(code)
	for i, char := range code {
		// 随机颜色（深色）
		textColor := color.RGBA{
			R: uint8(20 + r.Intn(100)),
			G: uint8(20 + r.Intn(100)),
			B: uint8(20 + r.Intn(100)),
			A: 255,
		}
		// 随机Y偏移
		yOffset := 10 + r.Intn(height/3)
		drawChar(img, char, i*charWidth+10, yOffset, textColor)
	}

	return img
}

// drawLine 画线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// drawChar 绘制字符（简单方块表示）
func drawChar(img *image.RGBA, char rune, x, y int, c color.Color) {
	// 简化版：用像素块表示字符
	pattern := getCharPattern(char)
	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == 1 {
				for py := 0; py < 3; py++ {
					for px := 0; px < 3; px++ {
						img.Set(x+dx*3+px, y+dy*3+py, c)
					}
				}
			}
		}
	}
}

// getCharPattern 获取字符的像素模式
func getCharPattern(char rune) [][]int {
	// 简化的 5x7 像素字符模式
	patterns := map[rune]string{
		'A': "01110_10001_10001_11111_10001_10001_10001",
		'B': "11110_10001_10001_11110_10001_10001_11110",
		'C': "01111_10000_10000_10000_10000_10000_01111",
		'D': "11110_10001_10001_10001_10001_10001_11110",
		'E': "11111_10000_10000_11110_10000_10000_11111",
		'F': "11111_10000_10000_11110_10000_10000_10000",
		'G': "01111_10000_10000_10111_10001_10001_01110",
		'H': "10001_10001_10001_11111_10001_10001_10001",
		'J': "00111_00001_00001_00001_10001_10001_01110",
		'K': "10001_10010_10100_11000_10100_10010_10001",
		'L': "10000_10000_10000_10000_10000_10000_11111",
		'M': "10001_11011_10101_10101_10001_10001_10001",
		'N': "10001_11001_10101_10011_10001_10001_10001",
		'P': "11110_10001_10001_11110_10000_10000_10000",
		'Q': "01110_10001_10001_10101_10010_01001_10110",
		'R': "11110_10001_10001_11110_10100_10010_10001",
		'S': "01111_10000_10000_01110_00001_00001_11110",
		'T': "11111_00100_00100_00100_00100_00100_00100",
		'U': "10001_10001_10001_10001_10001_10001_01110",
		'V': "10001_10001_10001_10001_10001_01010_00100",
		'W': "10001_10001_10001_10101_10101_11011_10001",
		'X': "10001_01010_00100_00100_00100_01010_10001",
		'Y': "10001_10001_01010_00100_00100_00100_00100",
		'Z': "11111_00001_00010_00100_01000_10000_11111",
		'2': "01110_10001_00001_00110_01000_10000_11111",
		'3': "11110_00001_00001_00110_00001_00001_11110",
		'4': "10001_10001_10001_11111_00001_00001_00001",
		'5': "11111_10000_11110_00001_00001_10001_01110",
		'6': "01110_10000_10000_11110_10001_10001_01110",
		'7': "11111_00001_00010_00100_00100_00100_00100",
		'8': "01110_10001_10001_01110_10001_10001_01110",
		'9': "01110_10001_10001_01111_00001_00001_01110",
	}

	var result [][]int
	if pattern, ok := patterns[char]; ok {
		rows := split(pattern, "_")
		for _, row := range rows {
			var rowPixels []int
			for _, c := range row {
				if c == '1' {
					rowPixels = append(rowPixels, 1)
				} else {
					rowPixels = append(rowPixels, 0)
				}
			}
			result = append(result, rowPixels)
		}
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}