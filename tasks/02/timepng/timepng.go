package timepng

import (
	"image"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"time"
)

// TimePNG записывает в `out` картинку в формате png с текущим временем
func TimePNG(out io.Writer, t time.Time, c color.Color, scale int) {
	img := buildTimeImage(t, c, scale)
	png.Encode(out, img)
}

// buildTimeImage создает новое изображение с временем `t`
func buildTimeImage(t time.Time, c color.Color, scale int) *image.RGBA {
	t_arr := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	x_size := scale * 3
	y_size := scale * 5
	img := image.NewRGBA(image.Rect(0, 0, (x_size + scale) * len(t_arr) - scale, y_size))

	mask := []int{}
	for y := 0; y < 5; y++ {
		for l := 0; l < len(t_arr); l++ {
			ch := rune(t_arr[l])
			if l != 0 {
				mask = append(mask, 0)
			}
			for x := 0; x < 3; x++ {
				mask = append(mask, nums[ch][y * 3 + x])
			}
		}
	}
	fillWithMask(img, mask, c, scale)
	return img
}

// fillWithMask заполняет изображение `img` цветом `c` по маске `mask`. Маска `mask`
// должна иметь пропорциональные размеры `img` с учетом фактора `scale`
// NOTE: Так как это вспомогательная функция, можно считать, что mask имеет размер (3x5)
func fillWithMask(img *image.RGBA, mask []int, c color.Color, scale int) {
	y_size := scale * 5
	for x := 0; x < len(mask) * scale / 5; x++ {
		i := x / scale
		for y := 0; y < y_size; y++ {
			j := y / scale
			if mask[j * len(mask) / 5 + i] == 1 {
				img.Set(x, y, c)
			}
		}
	}
}

var nums = map[rune][]int{
	'0': {
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'1': {
		0, 1, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
		0, 0, 1,
	},
	'2': {
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	'3': {
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	'4': {
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		0, 0, 1,
	},
	'5': {
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	'6': {
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'7': {
		1, 1, 1,
		0, 0, 1,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	},
	'8': {
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
	},
	'9': {
		1, 1, 1,
		1, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
	':': {
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
	},
}
