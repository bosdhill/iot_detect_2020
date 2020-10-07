package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"runtime"
	"sort"
	"time"
)

// TODO different color for each class -- can be used when augmenting images
//var colors = [20]color.RGBA{
//	color.RGBA{230, 25, 75, 0},
//	color.RGBA{60, 180, 75, 0},
//	color.RGBA{255, 225, 25, 0},
//	color.RGBA{0, 130, 200, 0},
//	color.RGBA{245, 130, 48, 0},
//	color.RGBA{145, 30, 180, 0},
//	color.RGBA{70, 240, 240, 0},
//	color.RGBA{240, 50, 230, 0},
//	color.RGBA{210, 245, 60, 0},
//	color.RGBA{250, 190, 190, 0},
//	color.RGBA{0, 128, 128, 0},
//	color.RGBA{230, 190, 255, 0},
//	color.RGBA{170, 110, 40, 0},
//	color.RGBA{255, 250, 200, 0},
//	color.RGBA{128, 0, 0, 0},
//	color.RGBA{170, 255, 195, 0},
//	color.RGBA{128, 128, 0, 0},
//	color.RGBA{255, 215, 180, 0},
//	color.RGBA{0, 0, 128, 0},
//	color.RGBA{128, 128, 128, 0},
//}

// Box represents the bounding box dimensions and class probabilities,
// confidence, and current class index
type Box struct {
	x               float32
	y               float32
	w               float32
	h               float32
	classProbs      [numClasses]float32
	confidence      float32
	currentClassIdx int
}

func regionLayer(predictions *gocv.Mat, transposePredictions bool, imgHeight, imgWidth float32) (map[string]bool, map[string]([]*BoundingBox)) {

	var data [w * h * 5 * (numClasses + 5)]float32
	var label string

	if transposePredictions {
		*predictions = predictions.Reshape(1, 425)
		data = transpose(predictions)
	} else {
		data = matToArray(predictions)
	}

	var boxes []Box
	for i := 0; i < numBoxes; i++ {
		index := i * size
		var n = i % n
		var row = float32((i / n) / h)
		var col = float32((i / n) % w)

		box := Box{}

		box.x = (col + logisticActivate(data[index+0])) / blockwd
		box.y = (row + logisticActivate(data[index+1])) / blockwd
		box.w = float32(math.Exp(float64(data[index+2]))) * anchors[2*n] / blockwd
		box.h = float32(math.Exp(float64(data[index+3]))) * anchors[2*n+1] / blockwd

		box.confidence = logisticActivate(data[index+4])

		if box.confidence < thresh {
			continue
		}

		box.classProbs = softmax(data[index+5 : index+5+numClasses])
		for j := 0; j < numClasses; j++ {
			box.classProbs[j] *= box.confidence
			if box.classProbs[j] < thresh {
				box.classProbs[j] = 0
			}
		}

		boxes = append(boxes, box)
	}

	// Non-Maximum-Suppression
	for k := 0; k < numClasses; k++ {
		for i := 0; i < len(boxes); i++ {
			boxes[i].currentClassIdx = k
		}

		sort.Sort(sort.Reverse(IndexSortList(boxes)))

		for i := 0; i < len(boxes); i++ {
			if boxes[i].classProbs[k] == 0 {
				continue
			}

			for j := i + 1; j < len(boxes); j++ {
				if boxIou(boxes[i], boxes[j]) > nmsThreshold {
					boxes[j].classProbs[k] = 0
				}
			}
		}
	}

	detections := make(map[string]([]*BoundingBox))
	labels := make(map[string]bool)

	for i := 0; i < len(boxes); i++ {
		maxI := maxIndex(boxes[i].classProbs[:])

		if maxI == -1 || boxes[i].classProbs[maxI] < thresh {
			continue
		}

		left := (boxes[i].x - boxes[i].w/2.) * imgWidth
		right := (boxes[i].x + boxes[i].w/2.) * imgWidth
		top := (boxes[i].y - boxes[i].h/2.) * imgHeight
		bottom := (boxes[i].y + boxes[i].h/2.) * imgHeight

		if left < 0 {
			left = 0
		}
		if right > imgWidth {
			right = imgWidth
		}
		if top < 0 {
			top = 0
		}
		if bottom > imgHeight {
			bottom = imgHeight
		}

		if left > right || top > bottom {
			continue
		}

		if int(right-left) == 0 || int(bottom-top) == 0 {
			continue
		}
		label = classNames[maxI]

		bbBox := BoundingBox{
			TopLeftX:     int(left),
			TopLeftY:     int(top),
			BottomRightX: int(right),
			BottomRightY: int(bottom),
			Confidence:   boxes[i].classProbs[maxI],
		}
		detections[label] = append(detections[label], &bbBox)

		if !labels[label] {
			labels[label] = true
		}
	}

	return labels, detections
}

func matToArray(m *gocv.Mat) [w * h * 5 * (numClasses + 5)]float32 {

	result := [w * h * 5 * (numClasses + 5)]float32{}
	i := 0
	for r := 0; r < m.Rows(); r++ {
		for c := 0; c < m.Cols(); c++ {
			result[i] = m.GetFloatAt(r, c)
			i++
		}
	}

	return result
}

func transpose(gocvMat *gocv.Mat) [w * h * 5 * (numClasses + 5)]float32 {

	result := [w * h * 5 * (numClasses + 5)]float32{}
	i := 0
	for c := 0; c < gocvMat.Cols(); c++ {
		for r := 0; r < gocvMat.Rows(); r++ {
			result[i] = gocvMat.GetFloatAt(r, c)
			i++
		}
	}

	return result
}

/*
 * Sorting intermediate results
 */

// IndexSortList is the sorted list of indices
type IndexSortList []Box

func (i IndexSortList) Len() int {
	return len(i)
}

func (i IndexSortList) Swap(j, k int) {
	i[j], i[k] = i[k], i[j]
}

func (i IndexSortList) Less(j, k int) bool {
	classIdx := i[j].currentClassIdx
	return i[j].classProbs[classIdx]-i[k].classProbs[classIdx] < 0
}

func logisticActivate(x float32) float32 {
	return 1.0 / (1.0 + float32(math.Exp(float64(-x))))
}

func softmax(x []float32) [numClasses]float32 {
	var sum float32 = 0.0
	var largest float32 = 0.0
	var e float32

	var output [numClasses]float32

	for i := 0; i < numClasses; i++ {
		if x[i] > largest {
			largest = x[i]
		}
	}

	for i := 0; i < numClasses; i++ {
		e = float32(math.Exp(float64(x[i] - largest)))
		sum += e
		output[i] = e
	}

	if sum > 1 {
		for i := 0; i < numClasses; i++ {
			output[i] /= sum
		}
	}

	return output
}

func overlap(x1, w1, x2, w2 float32) float32 {
	l1 := x1 - w1/2
	l2 := x2 - w2/2
	left := math.Max(float64(l1), float64(l2))

	r1 := x1 + w1/2
	r2 := x2 + w2/2
	right := math.Min(float64(r1), float64(r2))

	return float32(right - left)
}

func boxIntersection(a, b Box) float32 {
	w := overlap(a.x, a.w, b.x, b.w)
	h := overlap(a.y, a.h, b.y, b.h)
	if w < 0 || h < 0 {
		return 0
	}

	area := w * h
	return area
}

func boxUnion(a, b Box) float32 {
	i := boxIntersection(a, b)
	u := a.w*a.h + b.w*b.h - i
	return u
}

func boxIou(a, b Box) float32 {
	return boxIntersection(a, b) / boxUnion(a, b)
}

func maxIndex(a []float32) int {
	if len(a) == 0 {
		return -1
	}

	maxI := 0
	maxVal := math.Inf(-1)
	minVal := math.Inf(1)

	for i, val := range a {
		if float64(val) > maxVal {
			maxI = i
			maxVal = float64(val)
		}
		if float64(val) < minVal {
			minVal = float64(val)
		}
	}

	if maxVal == minVal {
		return -1
	}

	return maxI
}

// ObjectDetect contains the object detection model
type ObjectDetect struct {
	net  *gocv.Net
	eCtx *EdgeContext
}

// NewObjectDetection returns a new object detection component
func NewObjectDetection(eCtx *EdgeContext) (*ObjectDetect, error) {
	log.Println("NewObjectDetection")
	caffeNet := gocv.ReadNetFromCaffe(proto, model)
	if caffeNet.Empty() {
		return nil, fmt.Errorf("cannot read network model from: %v %v", proto, model)
	}
	return &ObjectDetect{net: &caffeNet, eCtx: eCtx}, nil
}

func (od *ObjectDetect) caffeWorker(imgChan chan *gocv.Mat, resChan chan DetectionResult) {
	log.Println("caffeWorker")
	sec := time.Duration(0)
	count := 0
	for img := range imgChan {
		if img.Empty() {
			log.Println("Img is Empty")
			continue
		}
		t := time.Now()
		blob := gocv.BlobFromImage(*img, 1.0/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
		od.net.SetInput(blob, "data")
		prob := od.net.Forward("conv9")
		probMat := prob.Reshape(1, 1)

		labels, labelBoxes := regionLayer(&probMat, true, float32(img.Rows()), float32(img.Cols()))
		//time.Sleep(time.Millisecond*30)
		e := time.Since(t)
		log.Println("detect time", e)
		sec += e
		count++
		log.Println("last AVG", sec/time.Duration(count))

		resChan <- DetectionResult{
			Empty:         len(labels) == 0,
			DetectionTime: time.Now().UnixNano(),
			Labels:        labels,
			Img:           img.Clone(),
			LabelBoxes:    labelBoxes,
		}

		blob.Close()
		probMat.Close()
		prob.Close()
		img.Close()
		runtime.GC()
		// close mat to avoid memory leak
		//if err := blob.Close(); err != nil {
		//	log.Fatalf("caffeWorker: Could not close blob mat with err = %s", err)
		//}
		//if err := probMat.Close(); err != nil {
		//	log.Fatalf("caffeWorker: Could not close probMat mat with err = %s", err)
		//}
		//if err := prob.Close(); err != nil {
		//	log.Fatalf("caffeWorker: Could not close prob mat with err = %s", err)
		//}
		//if err := img.Close(); err != nil {
		//	log.Fatalf("caffeWorker: Could not close img mat with err = %s", err)
		//}
		//log.Println("profile count:", gocv.MatProfile.Count())
		//var b bytes.Buffer
		//gocv.MatProfile.WriteTo(&b, 1)
		//log.Println("Mat frames", b.String())
	}
	close(resChan)
}
