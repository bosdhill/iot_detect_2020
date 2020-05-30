package main

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"log"
	"math"
	"sort"
	"time"
)

const numClasses = 80
const N = 5
const size = numClasses + N
const w = 12
const h = 12
const blockwd float32 = 13
const numBoxes = h*w*N
const thresh = 0.2
const nms_threshold = 0.4

var (
	proto = "model/tiny_yolo_deploy.prototxt"
	model = "model/tiny_yolo.caffemodel"
	classNames = [numClasses]string{"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train",
		"truck", "boat", "traffic light", "fire hydrant", "stop sign",
		"parking meter", "bench", "bird", "cat", "dog", "horse", "sheep", "cow",
		"elephant", "bear", "zebra", "giraffe", "backpack", "umbrella",
		"handbag", "tie", "suitcase", "frisbee", "skis", "snowboard",
		"sports ball", "kite", "baseball bat", "baseball glove", "skateboard",
		"surfboard", "tennis racket", "bottle", "wine glass", "cup", "fork",
		"knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange",
		"broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair",
		"couch", "potted plant", "bed", "dining table", "toilet", "tv",
		"laptop", "mouse", "remote", "keyboard", "cell phone", "microwave",
		"oven", "toaster", "sink", "refrigerator", "book", "clock", "vase",
		"scissors", "teddy bear", "hair drier", "toothbrush"}
	anchors = [2*N]float32{0.738768, 0.874946, 2.42204, 2.65704, 4.30971, 7.04493, 10.246, 4.59428, 12.6868, 11.8741}
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

type Box struct {
	x float32
	y float32
	w float32
	h float32
	classProbs [numClasses]float32
	confidence float32
	currentClassIdx int
}

func regionLayer(predictions gocv.Mat, transposePredictions bool, img_height, img_width float32) (map[string]bool, map[string]([]*BoundingBox)) {

	var data [w*h*5*(numClasses+5)]float32
	var label string

	if transposePredictions {
		predictions = predictions.Reshape(1, 425)
		data = transpose(&predictions)
	} else {
		data = matToArray(&predictions)
	}


	var boxes []Box
	for i := 0; i < numBoxes; i++ {
		index := i * size
		var n = i % N
		var row = float32((i/N) / h)
		var col = float32((i/N) % w)

		box := Box{}

		box.x = (col + logisticActivate(data[index + 0])) / blockwd
		box.y = (row + logisticActivate(data[index + 1])) / blockwd
		box.w = float32(math.Exp(float64(data[index + 2]))) * anchors[2*n] / blockwd
		box.h = float32(math.Exp(float64(data[index + 3]))) * anchors[2*n+1] / blockwd

		box.confidence = logisticActivate(data[index + 4])

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

			for j := i+1; j < len(boxes); j++ {
				if box_iou(boxes[i], boxes[j]) > nms_threshold {
					boxes[j].classProbs[k] = 0
				}
			}
		}
	}

	detections := make(map[string]([]*BoundingBox))
	labels := make(map[string]bool)

	for i := 0; i < len(boxes); i++ {
		max_i := max_index(boxes[i].classProbs[:])

		if max_i == -1 || boxes[i].classProbs[max_i] < thresh {
			continue
		}

		left := (boxes[i].x - boxes[i].w/2.) * img_width
		right := (boxes[i].x + boxes[i].w/2.) * img_width
		top := (boxes[i].y - boxes[i].h/2.) * img_height
		bottom := (boxes[i].y + boxes[i].h/2.) * img_height

		if left < 0 { left = 0 }
		if right > img_width { right = img_width }
		if top < 0 { top = 0 }
		if bottom > img_height { bottom = img_height }


		if left > right || top > bottom {
			continue
		}

		if int(right - left) == 0 || int(bottom - top) == 0 {
			continue
		}
		label = classNames[max_i]

		bbBox := BoundingBox{
			TopLeftX:     int(left),
			TopLeftY:     int(top),
			BottomRightX: int(right),
			BottomRightY: int(bottom),
			Confidence:   boxes[i].classProbs[max_i],
		}
		detections[label] = append(detections[label], &bbBox)

		if !labels[label] {
			labels[label] = true
		}
	}

	return labels, detections
}

func matToArray(m *gocv.Mat) [w*h*5*(numClasses+5)]float32 {

	result := [w*h*5*(numClasses+5)]float32{}
	i := 0
	for r := 0; r < m.Rows(); r++ {
		for c := 0; c < m.Cols(); c++ {
			result[i] = m.GetFloatAt(r, c)
			i++
		}
	}

	return result
}


func transpose(gocvMat *gocv.Mat) [w*h*5*(numClasses+5)]float32 {

	result := [w*h*5*(numClasses+5)]float32{}
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


type IndexSortList []Box

func (i IndexSortList) Len() int {
	return len(i)
}

func (i IndexSortList) Swap(j,k int)  {
	i[j], i[k] = i[k], i[j]
}

func (i IndexSortList) Less(j,k int) bool  {
	classIdx := i[j].currentClassIdx
	return i[j].classProbs[classIdx] - i[k].classProbs[classIdx] < 0
}

func logisticActivate(x float32) float32 {
	return 1.0/(1.0 + float32(math.Exp(float64(-x))))
}


func softmax(x []float32) [numClasses]float32 {
	var sum float32 = 0.0
	var largest float32 = 0.0
	var e float32

	var output [numClasses]float32

	for i:=0; i<numClasses; i++ {
		if x[i] > largest {
			largest = x[i]
		}
	}

	for i:=0; i<numClasses; i++ {
		e = float32(math.Exp(float64(x[i] - largest)))
		sum += e
		output[i] = e
	}

	if sum > 1 {
		for i:=0; i<numClasses; i++ {
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


func box_intersection(a, b Box) float32 {
	w := overlap(a.x, a.w, b.x, b.w);
	h := overlap(a.y, a.h, b.y, b.h);
	if w < 0 || h < 0 {
		return 0
	}

	area := w*h
	return area
}

func box_union(a,b Box) float32 {
	i := box_intersection(a, b)
	u := a.w*a.h + b.w*b.h - i
	return u
}


func box_iou(a, b Box) float32 {
	return box_intersection(a,b) / box_union(a,b)
}

func max_index(a []float32) int {
	if len(a) == 0 {
		return -1
	}

	max_i := 0
	max_val := math.Inf(-1)
	min_val := math.Inf(1)

	for i, val := range (a) {
		if float64(val) > max_val {
			max_i = i
			max_val = float64(val)
		}
		if float64(val) < min_val {
			min_val = float64(val)
		}
	}

	if max_val == min_val {
		return -1
	}

	return max_i
}

type objectDetect struct {
	net *gocv.Net
	eCtx *EdgeContext
}

func NewObjectDetection(eCtx *EdgeContext) (*objectDetect, error){
	log.Println("NewObjectDetection")
	caffeNet := gocv.ReadNetFromCaffe(proto, model)
	if caffeNet.Empty() {
		return nil, errors.New(fmt.Sprintf("Error reading network model from : %v %v\n", proto, model))
	}
	return &objectDetect{net: &caffeNet, eCtx: eCtx}, nil
}

func (od *objectDetect) caffeWorker(imgChan chan *gocv.Mat, resChan chan *DetectionResult) {
	log.Println("caffeWorker")
	img := gocv.NewMat()
	defer img.Close()

	blob := gocv.NewMat()
	defer blob.Close()

	prob := gocv.NewMat()
	defer prob.Close()

	probMat := gocv.NewMat()
	defer probMat.Close()

	sec := time.Duration(0)
	count := 0
	for item := range imgChan {
		if item.Empty(){
			log.Println("Img is Empty")
			continue
		}
		t := time.Now()
		img = item.Clone()
		blob = gocv.BlobFromImage(img, 1.0/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
		od.net.SetInput(blob, "data")
		prob = od.net.Forward("conv9")
		probMat := prob.Reshape(1,1)

		labels, labelBoxes := regionLayer(probMat, true, float32(img.Rows()), float32(img.Cols()))
		//time.Sleep(time.Millisecond*30)
		e := time.Since(t)
		log.Println("detect time", e)
		sec += e
		count++
		log.Println("last AVG", sec / time.Duration(count))


		resChan <- &DetectionResult{
			Empty:         len(labels) == 0,
			DetectionTime: time.Now().UnixNano(),
			Labels:        labels,
			Img:           img.Clone(),
			LabelBoxes:    labelBoxes}
	}
	close(resChan)
}
