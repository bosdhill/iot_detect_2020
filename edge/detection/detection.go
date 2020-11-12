// Ref: https://github.com/dymat/GOLOv2
package detection

import (
	"context"
	"fmt"
	aod "github.com/bosdhill/iot_detect_2020/edge/actionondetect"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"math"
	"runtime"
	"sort"
	"time"
)

// NumClasses used in object detection
const NumClasses = 80
const n = 5
const size = NumClasses + n
const w = 12
const h = 12
const blockwd float32 = 13
const numBoxes = h * w * n
const thresh = 0.2
const nmsThreshold = 0.4

var (
	classNames = [NumClasses]string{"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train",
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
	anchors = [2 * n]float32{0.738768, 0.874946, 2.42204, 2.65704, 4.30971, 7.04493, 10.246, 4.59428, 12.6868, 11.8741}

	// ClassNames is a map of the classnames to pass to RegisterEvents
	ClassNames = map[string]bool{
		"person":         true,
		"bicycle":        true,
		"car":            true,
		"motorcycle":     true,
		"airplane":       true,
		"bus":            true,
		"train":          true,
		"truck":          true,
		"boat":           true,
		"traffic light":  true,
		"fire hydrant":   true,
		"stop sign":      true,
		"parking meter":  true,
		"bench":          true,
		"bird":           true,
		"cat":            true,
		"dog":            true,
		"horse":          true,
		"sheep":          true,
		"cow":            true,
		"elephant":       true,
		"bear":           true,
		"zebra":          true,
		"giraffe":        true,
		"backpack":       true,
		"umbrella":       true,
		"handbag":        true,
		"tie":            true,
		"suitcase":       true,
		"frisbee":        true,
		"skis":           true,
		"snowboard":      true,
		"sports ball":    true,
		"kite":           true,
		"baseball bat":   true,
		"baseball glove": true,
		"skateboard":     true,
		"surfboard":      true,
		"tennis racket":  true,
		"bottle":         true,
		"wine glass":     true,
		"cup":            true,
		"fork":           true,
		"knife":          true,
		"spoon":          true,
		"bowl":           true,
		"banana":         true,
		"apple":          true,
		"sandwich":       true,
		"orange":         true,
		"broccoli":       true,
		"carrot":         true,
		"hot dog":        true,
		"pizza":          true,
		"donut":          true,
		"cake":           true,
		"chair":          true,
		"couch":          true,
		"potted plant":   true,
		"bed":            true,
		"dining table":   true,
		"toilet":         true,
		"tv":             true,
		"laptop":         true,
		"mouse":          true,
		"remote":         true,
		"keyboard":       true,
		"cell phone":     true,
		"microwave":      true,
		"oven":           true,
		"toaster":        true,
		"sink":           true,
		"refrigerator":   true,
		"book":           true,
		"clock":          true,
		"vase":           true,
		"scissors":       true,
		"teddy bear":     true,
		"hair drier":     true,
		"toothbrush":     true}

	colors = [20]color.RGBA{
		color.RGBA{230, 25, 75, 0},
		color.RGBA{60, 180, 75, 0},
		color.RGBA{255, 225, 25, 0},
		color.RGBA{0, 130, 200, 0},
		color.RGBA{245, 130, 48, 0},
		color.RGBA{145, 30, 180, 0},
		color.RGBA{70, 240, 240, 0},
		color.RGBA{240, 50, 230, 0},
		color.RGBA{210, 245, 60, 0},
		color.RGBA{250, 190, 190, 0},
		color.RGBA{0, 128, 128, 0},
		color.RGBA{230, 190, 255, 0},
		color.RGBA{170, 110, 40, 0},
		color.RGBA{255, 250, 200, 0},
		color.RGBA{128, 0, 0, 0},
		color.RGBA{170, 255, 195, 0},
		color.RGBA{128, 128, 0, 0},
		color.RGBA{255, 215, 180, 0},
		color.RGBA{0, 0, 128, 0},
		color.RGBA{128, 128, 128, 0},
	}
)

// Box represents the bounding box dimensions and class probabilities,
// confidence, and current class index
type Box struct {
	x               float32
	y               float32
	w               float32
	h               float32
	classProbs      [NumClasses]float32
	confidence      float32
	currentClassIdx int
}

func regionLayer(predictions *gocv.Mat, transposePredictions bool, imgHeight, imgWidth float32) ([]string, map[string]int32, map[string]*pb.BoundingBoxes) {

	var data [w * h * 5 * (NumClasses + 5)]float32
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
		var N = i % n
		var row = float32((i / n) / h)
		var col = float32((i / n) % w)

		box := Box{}

		box.x = (col + logisticActivate(data[index+0])) / blockwd
		box.y = (row + logisticActivate(data[index+1])) / blockwd
		box.w = float32(math.Exp(float64(data[index+2]))) * anchors[2*N] / blockwd
		box.h = float32(math.Exp(float64(data[index+3]))) * anchors[2*N+1] / blockwd
		box.confidence = logisticActivate(data[index+4])

		if box.confidence < thresh {
			continue
		}

		box.classProbs = softmax(data[index+5 : index+5+NumClasses])
		for j := 0; j < NumClasses; j++ {
			box.classProbs[j] *= box.confidence
			if box.classProbs[j] < thresh {
				box.classProbs[j] = 0
			}
		}

		boxes = append(boxes, box)
	}

	// Non-Maximum-Suppression
	for k := 0; k < NumClasses; k++ {
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

	boxesMap := make(map[string]*pb.BoundingBoxes)
	labelMap := make(map[string]int32)
	labels := make([]string, 0)

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

		bbBox := pb.BoundingBox{
			TopLeftX:     int32(left),
			TopLeftY:     int32(top),
			BottomRightX: int32(right),
			BottomRightY: int32(bottom),
			Confidence:   boxes[i].classProbs[maxI],
		}

		if boxesMap[label] == nil {
			boxesMap[label] = &pb.BoundingBoxes{
				LabelBoxes: []*pb.BoundingBox{},
			}
		}
		boxesMap[label].LabelBoxes = append(boxesMap[label].LabelBoxes, &bbBox)
		labelMap[label] += 1
		labels = append(labels, label)
	}

	return labels, labelMap, boxesMap
}

func matToArray(m *gocv.Mat) [w * h * 5 * (NumClasses + 5)]float32 {

	result := [w * h * 5 * (NumClasses + 5)]float32{}
	i := 0
	for r := 0; r < m.Rows(); r++ {
		for c := 0; c < m.Cols(); c++ {
			result[i] = m.GetFloatAt(r, c)
			i++
		}
	}

	return result
}

func transpose(gocvMat *gocv.Mat) [w * h * 5 * (NumClasses + 5)]float32 {

	result := [w * h * 5 * (NumClasses + 5)]float32{}
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

func softmax(x []float32) [NumClasses]float32 {
	var sum float32 = 0.0
	var largest float32 = 0.0
	var e float32

	var output [NumClasses]float32

	for i := 0; i < NumClasses; i++ {
		if x[i] > largest {
			largest = x[i]
		}
	}

	for i := 0; i < NumClasses; i++ {
		e = float32(math.Exp(float64(x[i] - largest)))
		sum += e
		output[i] = e
	}

	if sum > 1 {
		for i := 0; i < NumClasses; i++ {
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

// imgToMat handles deserializing the pb.Image to a gocv.Mat
func imgToMat(img *pb.Image) *gocv.Mat {
	height := int(img.Rows)
	width := int(img.Cols)
	mType := gocv.MatType(img.Type)
	mat, err := gocv.NewMatFromBytes(height, width, mType, img.Image)
	if mType != gocv.MatTypeCV32F {
		mat.ConvertTo(&mat, gocv.MatTypeCV32F)
	}
	if err != nil {
		log.Fatal(err)
	}
	return &mat
}

// ObjectDetect contains the object detection model
type ObjectDetect struct {
	net *gocv.Net
	ctx context.Context
	aod *aod.ActionOnDetect
}

// NewObjectDetection returns a new object detection component
func NewObjectDetection(ctx context.Context, aod *aod.ActionOnDetect, withCuda bool, proto, model string) (*ObjectDetect, error) {
	log.Println("NewObjectDetection")
	caffeNet := gocv.ReadNetFromCaffe(proto, model)
	if caffeNet.Empty() {
		return nil, fmt.Errorf("cannot read network model from: %v %v", proto, model)
	}
	// Set net backend type as CUDA if running on the Jetson Nano
	if withCuda {
		log.Println("Built with CUDA backend enabled")
		caffeNet.SetPreferableBackend(gocv.NetBackendType(gocv.NetBackendCUDA))
		caffeNet.SetPreferableTarget(gocv.NetTargetType(gocv.NetTargetCUDA))
	}
	return &ObjectDetect{net: &caffeNet, ctx: ctx, aod: aod}, nil
}

func (od *ObjectDetect) CaffeWorker(imgChan chan *pb.Image, drCh chan pb.DetectionResult) {
	log.Println("caffeWorker")
	sec := time.Duration(0)
	count := 0
	for img := range imgChan {
		mat := imgToMat(img)
		if mat.Empty() {
			log.Println("Img is Empty")
			continue
		}
		t := time.Now()

		//gocv.IMWrite(fmt.Sprintf("./temp/received_pre_%v.jpeg", count), *mat)
		//gocv.Resize(*mat, mat, image.Pt(416, 416), 1.0/255.0, 1.0/255.0, gocv.InterpolationDefault)
		//gocv.IMWrite(fmt.Sprintf("./temp/received_%v.jpeg", count), *mat)
		blob := gocv.BlobFromImage(*mat, 1.0/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false)
		od.net.SetInput(blob, "data")

		// In gocv v0.25.0, the name of the last output layer is lost, so its replaced with the name of top to retrieve
		// the output blob.
		// Ref: https://stackoverflow.com/a/45653638/10741562
		prob := od.net.Forward("result")
		probMat := prob.Reshape(1, 1)

		labels, labelMap, labelBoxes := regionLayer(&probMat, true, float32(mat.Rows()), float32(mat.Cols()))

		e := time.Since(t)
		log.Println("detect time", e)
		sec += e
		count++
		log.Println("last AVG", sec/time.Duration(count))

		dr := pb.DetectionResult{
			Empty:         len(labels) == 0,
			DetectionTime: time.Now().UnixNano(),
			LabelMap:      labelMap,
			Labels:        labels,
			Img:           img,
			LabelBoxes:    labelBoxes,
		}

		log.Printf("detected labelsMap: %v\n", labelMap)

		if err := blob.Close(); err != nil {
			log.Println("blob close error: ", err)
		}
		if err := probMat.Close(); err != nil {
			log.Println("probMat close error: ", err)
		}
		if err := prob.Close(); err != nil {
			log.Println("prob close error: ", err)
		}
		if err := mat.Close(); err != nil {
			log.Println("mat close error: ", err)
		}
		runtime.GC()
		od.aod.CheckEvents(&dr)
		drCh <- dr

		//if *matprofile {
		//	log.Println("profile count:", gocv.MatProfile.Count())
		//	var b bytes.Buffer
		//	gocv.MatProfile.WriteTo(&b, 1)
		//	log.Println("Mat frames", b.String())
		//}
	}
	close(imgChan)
}
