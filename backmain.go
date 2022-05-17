package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func rootHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello from Go Echo")
}

func CalcGR(img *image.RGBA, rmin float64, rmax float64, imin float64, imax float64, offset int, nGR int, fchannel chan int) {
	for y := 0; y < 800; y++ {
		for x := offset; x < 800; x += nGR {
			var C = complex(rmin+(rmax-rmin)*float64(x)/800, imin+(imax-imin)*float64(y)/800)
			var V = C
			var it uint8 = 0
			for ; it <= 100; it++ {
				V = V*V + C
				if cmplx.Abs(V) > 2 {
					break
				}
			}
			img.Set(x, y, color.RGBA{uint8(cmplx.Abs(V) * 40), uint8(cmplx.Abs(V) * 120), it * 2, 255})
		}
	}
	fchannel <- 1
}

func getMapHandler(c echo.Context) error {
	realMin, err1 := strconv.ParseFloat(c.FormValue("realmin"), 64)
	realMax, err2 := strconv.ParseFloat(c.FormValue("realmax"), 64)
	imaginaryMin, err3 := strconv.ParseFloat(c.FormValue("imaginarymin"), 64)
	imaginaryMax, err4 := strconv.ParseFloat(c.FormValue("imaginarymax"), 64)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil ||
		realMin >= realMax || imaginaryMin >= imaginaryMax {
		return c.String(http.StatusBadRequest, "incorrect parameter")
	}
	img := image.NewRGBA(image.Rect(0, 0, 800, 800))
	nGR := 4
	fchannel := make(chan int)
	defer close(fchannel)
	r := make([]int, nGR)
	for i := range r {
		go CalcGR(img, realMin, realMax, imaginaryMin, imaginaryMax, i, nGR, fchannel)
	}
	for range r {
		<-fchannel
	}

	var buff bytes.Buffer
	png.Encode(&buff, img)
	encodedString := base64.StdEncoding.EncodeToString(buff.Bytes())
	return c.String(http.StatusOK, encodedString)
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.GET("/", rootHandler)
	e.GET("/GetMap", getMapHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
