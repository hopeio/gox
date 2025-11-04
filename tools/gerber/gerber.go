package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"unsafe"

	"github.com/hopeio/gox/encoding/gerber"
	"github.com/hopeio/gox/log"
	"github.com/hopeio/gox/math/geom"
	imagex "github.com/hopeio/gox/media/image"
	gocvx "github.com/hopeio/gox/sdk/gocv"
	"gocv.io/x/gocv"
)

func main() {
	path := `D:\Gerber_TopLayer.GTL`
	maxWidth := 100.0
	maxHeight := 170.0
	rotation := 0.0
	p := &gerber.StoreProcessor{}
	pp := gerber.NewParser(p)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	err = pp.Parse(f)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	hashInBytes := hash.Sum(nil)[:16]
	binPath, _ := os.Executable()
	dir := filepath.Dir(binPath)
	output := filepath.Join(dir, hex.EncodeToString(hashInBytes)+".png")
	file.Close()
	_, err = os.Stat(output)
	factor := float64(40)
	if os.IsNotExist(err) {
		err = exec.Command("gerbv.exe", "-a", "-x", "png", "-B", "0", "-D", strconv.Itoa(int(factor)/10*254), "-o", output, path).Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	dcountMap := make(map[float64]int)
	for _, circle := range p.Circles {
		dcountMap[circle.Diameter]++
	}
	var maxCount int
	var maxCountRadius float64
	for d, count := range dcountMap {
		if count > maxCount {
			maxCount = count
			maxCountRadius = d / 2
		}
	}
	// 1mm mark更常见
	if dcountMap[1] >= 4 {
		maxCount = dcountMap[1]
		maxCountRadius = 0.5
	}

	rects, circles, imageRect := CvGerber(output, int(math.Round(maxCountRadius*factor)), int(math.Round(maxWidth*factor)), int(math.Round(maxHeight*factor)))
	log.Debugf("rects %d, circles: %d, radius: %d img size %v", len(rects), len(circles), int(math.Round(maxCountRadius*factor)), imageRect)

	centerX, centerY := float64(imageRect.Dx())/(2*factor), float64(imageRect.Dy())/(2*factor)
	newCenterX, newCenterY := float64(imageRect.Dy())/(2*factor), float64(imageRect.Dx())/(2*factor)
	affineMatrix := geom.NewTranslateRotationMat(geom.Pt(centerX, centerY), geom.Pt(newCenterX, newCenterY), rotation)

	//gerberCircle := p.Circles
	p.Circles = nil
	for _, circle := range circles {
		if rotation != 0 {
			newP := affineMatrix.Transform(geom.Pt(float64(circle.Center.X)/factor, float64(circle.Center.Y)/factor))
			p.Circle(&gerber.Circle{Circle: geom.Circle{newP, float64(circle.Center.Y) / (factor / 2)}})
		} else {
			p.Circle(&gerber.Circle{Circle: geom.Circle{geom.Pt(float64(circle.Center.X)/factor, float64(circle.Center.Y)/factor), float64(circle.Radius) / (factor / 2)}})
		}
	}
	var removeRectCount int
	p.Rects = nil
	p.Obrounds = nil
	for _, rect := range rects {
		bounds := rect.BoundingRect
		if bounds.Min.X < imageRect.Min.X || bounds.Min.Y < imageRect.Min.Y || bounds.Max.X > imageRect.Max.X || bounds.Max.Y > imageRect.Max.Y {
			continue
		}
		if removeRectCount < maxCount && slices.ContainsFunc(circles, func(c *imagex.Circle) bool {
			return (bounds.Min.X-imageRect.Min.X+bounds.Max.X-imageRect.Min.X)/2 >= c.Center.X-c.Radius &&
				(bounds.Min.X-imageRect.Min.X+bounds.Max.X-imageRect.Min.X)/2 <= c.Center.X+c.Radius &&
				(bounds.Min.Y-imageRect.Min.Y+bounds.Max.Y-imageRect.Min.Y)/2 >= c.Center.Y-c.Radius &&
				(bounds.Min.Y-imageRect.Min.Y+bounds.Max.Y-imageRect.Min.Y)/2 <= c.Center.Y+c.Radius
		}) {
			removeRectCount++
			continue
		}
		if rotation != 0 {
			newP := affineMatrix.Transform(geom.Pt(float64(rect.Center.X-imageRect.Min.X)/factor, float64(rect.Center.Y-imageRect.Min.Y)/factor))
			p.Rectangle(&gerber.Rectangle{Rectangle: geom.Rectangle{newP, float64(rect.Width) / factor, float64(rect.Height) / factor, rect.Angle + 90}})
		} else {
			p.Rectangle(&gerber.Rectangle{Rectangle: geom.Rectangle{geom.Pt(float64(rect.Center.X-imageRect.Min.X)/factor, float64(rect.Center.Y-imageRect.Min.Y)/factor), float64(rect.Width) / factor, float64(rect.Height) / factor, rect.Angle}})
		}
	}
}

type RotatedRect struct {
	Points       []image.Point
	BoundingRect image.Rectangle
	Center       image.Point
	Width        int
	Height       int
	Angle        float64
}

func CvGerber(path string, radius int, maxWidth, maxHeight int) ([]*RotatedRect, []*imagex.Circle, image.Rectangle) {
	var rects []*RotatedRect
	img := gocv.IMRead(path, gocv.IMReadUnchanged)
	defer img.Close()
	imgWidth, imgHeight := img.Cols(), img.Rows()
	gary := gocv.NewMat()
	defer gary.Close()
	gocv.CvtColor(img, &gary, gocv.ColorBGRToGray)
	binary := gocv.NewMat()
	defer binary.Close()
	gocv.Threshold(gary, &binary, 1, 255, gocv.ThresholdBinary)
	hierarchy := gocv.NewMat()
	defer hierarchy.Close()
	contours := gocv.FindContoursWithParams(binary, &hierarchy, gocv.RetrievalList, gocv.ChainApproxSimple)
	maxArea := 0
	maxRect := image.Rect(0, 0, 0, 0)
	secondRect := image.Rect(0, 0, 0, 0)
	for i := range contours.Size() {
		minRect := gocv.MinAreaRect(contours.At(i))
		rect := minRect.BoundingRect
		if rect.Min.X < 0 {
			rect.Min.X = 0
		}
		if rect.Max.X > imgWidth {
			rect.Max.X = imgWidth
		}
		if rect.Min.Y < 0 {
			rect.Min.Y = 0
		}
		if rect.Max.Y > imgHeight {
			rect.Max.Y = imgHeight
		}
		w, h := rect.Dx(), rect.Dy()
		area := w * h
		if area > maxArea {
			maxArea = area
			secondRect = maxRect
			maxRect = rect
		}
		notZero := gocv.CountNonZero(binary.Region(rect))
		fmt.Println(notZero)
		notZero = gocvx.CountNonZeroInPointsVector(binary, gocvx.PointsVectorFromPoints(minRect.Points))
		fmt.Println(notZero)
		fillRadio := float64(notZero) / float64(minRect.Height*minRect.Width)
		if w == 1 || h == 1 || fillRadio < 0.5 {
			continue
		}
		rects = append(rects, (*RotatedRect)(unsafe.Pointer(&minRect)))
	}

	if secondRect.Dx() < maxWidth-400 || secondRect.Dy() < maxHeight-400 {
		secondRect = image.Rect(0, 0, img.Cols(), img.Rows())
	}

	targetB := binary.Region(secondRect)

	blurred := gocv.NewMat()
	defer blurred.Close()
	gocv.GaussianBlur(targetB, &blurred, image.Pt(9, 9), 0, 0, gocv.BorderDefault)
	circleMat := gocv.NewMat()
	defer circleMat.Close()
	gocv.HoughCirclesWithParams(blurred, &circleMat, gocv.HoughGradient, 1, float64(radius*10), 30, 30, radius, radius)
	var circles []*imagex.Circle
	for i := range circleMat.Cols() {
		v := circleMat.GetVecfAt(0, i)
		x := int(v[0])
		y := int(v[1])
		r := int(v[2])
		area := (r * 2) * (r * 2)
		region := targetB.Region(image.Rect(max(x-r, 0), max(y-r, 0), x+r, y+r))
		pixels := gocv.CountNonZero(region)
		fillRatio := float64(pixels) / float64(area)
		if fillRatio >= 0.7 {
			circles = append(circles, &imagex.Circle{image.Pt(x, y), r})
		}
	}
	//debug
	for _, rect := range rects {
		gocv.Polylines(&img, gocv.NewPointsVectorFromPoints([][]image.Point{rect.Points}), true, color.RGBA{0, 255, 0, 0}, 1)
	}
	target := img.Region(secondRect)
	for _, c := range circles {
		gocv.Circle(&target, image.Pt(c.Center.X, c.Center.Y), c.Radius, color.RGBA{255, 0, 0, 0}, 1)
	}
	gocv.IMWrite("debug.png", target)
	return rects, circles, secondRect
}
