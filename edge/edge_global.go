package main

const numClasses = 80
const n = 5
const size = numClasses + n
const w = 12
const h = 12
const blockwd float32 = 13
const numBoxes = h * w * n
const thresh = 0.2
const nmsThreshold = 0.4

var (
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
	classNamesMap = map[string]bool{
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
	anchors = [2 * n]float32{0.738768, 0.874946, 2.42204, 2.65704, 4.30971, 7.04493, 10.246, 4.59428, 12.6868, 11.8741}
)
