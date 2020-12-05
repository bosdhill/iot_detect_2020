package realtimefilter

import (
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

const (
	notWrappedError    = "createLogicalQuery: query is not wrapped with $and or $or"
	notBsonDError      = "createLogicalQuery: pb.EventFiller.Filter is not bsonD"
	notBsonAError      = "createArrayQuery: pb.EventFiller.Filter inner is not bsonA"
	notAllError        = "createArrayQuery: pb.EventFiller.Filter is not wrapped with $all"
	LessThan           = "$lt"
	GreaterThan        = "$gt"
	Equal              = "$eq"
	GreaterThanOrEqual = "$gte"
	LessThanOrEqual    = "$lte"
	AndOp              = "$and"
	OrOp               = "$or"
	AllOp              = "$all"
)

type Set []realTimeFilter

// logicalQuery maps the logical query type to a map of the label to a comparison query
type logicalQuery map[string]map[string]bson.D

// query + pb.DetectionResult + pb.EventFilter = Event to send back to client
type realTimeFilter struct {
	query       interface{}
	eventFilter *pb.EventFilter
}

// New returns a slice of eventLabelPairs
func New(events *pb.EventFilters) (*Set, error) {
	log.Println("NewRealTimeFilter")
	set := make(Set, 0)
	for _, event := range events.GetEventFilters() {
		bFilter := event.GetFilter()

		// FIXME should be unmarshalled as bson.Raw for easy parsing.
		filter, err := UnmarshallBsonDFilter(bFilter)
		if err != nil {
			return nil, err
		}

		// check if its an $or statement
		logicalQuery, err := createLogicalQuery(filter, OrOp)
		if err == nil {
			set = append(set, realTimeFilter{logicalQuery, event})
		} else {
			// check if its an $and statement
			logicalQuery, err = createLogicalQuery(filter, AndOp)
			if err == nil {
				set = append(set, realTimeFilter{logicalQuery, event})
			} else {
				// finally, check if its an $all statement
				arrayQuery, err := createArrayQuery(filter, "labels", AllOp)
				if err != nil {
					return nil, err
				}
				set = append(set, realTimeFilter{arrayQuery, event})
			}
		}
	}
	return &set, nil
}

// createLogicalQuery takes the unmarshalled bson.D query from a pb.EventFilter and creates a query
func createLogicalQuery(filter bson.D, operator string) (logicalQuery, error) {
	//log.Println("createLogicalQuery")
	f := make(logicalQuery)
	m, ok := filter.Map()[operator]
	if !ok {
		return nil, fmt.Errorf(notWrappedError)
	}
	b, ok := m.(bson.D)
	if !ok {
		return nil, fmt.Errorf(notBsonDError)
	}
	f[operator] = make(map[string]bson.D)
	for label, elt := range b.Map() {
		b, ok := elt.(bson.D)
		if !ok {
			return nil, fmt.Errorf(notBsonDError)
		}
		f[operator][label] = b
	}
	//log.Println("Created logicalQuery:", f)
	return f, nil
}

// createArrayQuery takes the unmarshalled bson.D query from a pb.EventFilter and creates a bson.A for subset
// querying
func createArrayQuery(filter bson.D, array string, operator string) (bson.A, error) {
	//log.Println("createArrayQuery")
	m, ok := filter.Map()[array]
	if !ok {
		return nil, fmt.Errorf(notWrappedError)
	}
	b, ok := m.(bson.D)
	if !ok {
		return nil, fmt.Errorf(notBsonDError)
	}
	a, ok := b.Map()[operator]
	if !ok {
		return nil, fmt.Errorf(notAllError)
	}
	aSl, ok := a.(bson.A)
	if !ok {
		return nil, fmt.Errorf(notBsonAError)
	}
	//log.Println("Created arrayQuery:", aSl)
	return aSl, nil
}

// UnmarshallBsonDFilter unmarshals the query in the EventFilter and returns a query if it's valid, otherwise an error
func UnmarshallBsonDFilter(bFilter []byte) (bson.D, error) {
	var filter bson.D
	err := bson.Unmarshal(bFilter, &filter)
	if err != nil {
		return nil, err
	}
	return filter, nil
}

// UnmarshallBsonEFilter unmarshals the query in the EventFilter and returns a query if it's valid, otherwise an error
func UnmarshallBsonEFilter(bFilter []byte) (bson.E, error) {
	var filter bson.E
	err := bson.Unmarshal(bFilter, &filter)
	if err != nil {
		return bson.E{}, err
	}
	return filter, nil
}

// compare compares the number of labels to the value in the compareQuery
func compare(n int32, compareQuery bson.D) bool {
	op := compareQuery.Map()["key"].(string)
	bound := compareQuery.Map()["value"].(int32)

	switch op {
	case LessThan:
		return n < bound
	case LessThanOrEqual:
		return n <= bound
	case Equal:
		return n == bound
	case GreaterThan:
		return n > bound
	case GreaterThanOrEqual:
		return n >= bound
	}
	return false
}

// containsAll checks whether the keys of detectedLabels contains all labels
func containsAll(detectedLabels map[string]int32, labels bson.A) bool {
	//log.Println("containsAll")
	for _, label := range labels {
		_, ok := detectedLabels[label.(string)]
		if !ok {
			return false
		}
	}
	return true
}

// compareOr returns the "or" of the results of each compare query
func compareOr(detectedLabels map[string]int32, labelQuery map[string]bson.D) bool {
	//log.Println("compareOr")
	ret := false
	for label, compareQuery := range labelQuery {
		n, ok := detectedLabels[label]
		if ok {
			//log.Println("compare", n, compareQuery, compare(n, compareQuery))
			ret = ret || compare(n, compareQuery)
		}
	}
	log.Println("ret", ret)
	return ret
}

// compareAnd returns the "and" of the results of each compare query
func compareAnd(detectedLabels map[string]int32, labelQuery map[string]bson.D) bool {
	//log.Println("compareAnd")
	ret := true
	for label, compareQuery := range labelQuery {
		n, ok := detectedLabels[label]
		if !ok {
			return false
		}
		//log.Println("compare", n, compareQuery, compare(n, compareQuery))
		ret = ret && compare(n, compareQuery)
	}
	log.Println("ret", ret)
	return ret
}

// NewEvent returns a new Event based on the EventFilter and DetectionResult
func NewEvent(eventFilter *pb.EventFilter, dr pb.DetectionResult) *pb.Event {
	log.Println("NewEvent")
	// TODO omit DetectionResult fields based on flags in eventFilter
	return &pb.Event{
		Name:            eventFilter.Name,
		DetectionResult: &dr,
		AnnotatedImg:    nil,
	}
}

// NewEvents returns a set of Events for an EventFilter
func NewEvents(eventFilter *pb.EventFilter, drSl []pb.DetectionResult) []*pb.Event {
	log.Println("NewEvents")
	var events []*pb.Event
	for _, dr := range drSl {
		log.Println("Adding: ", dr.LabelNumber, dr.DetectionTime)
		events = append(events, NewEvent(eventFilter, dr))
	}
	return events
}

// GetEvents returns the all the Events for the EventFilters that the DetectionResult satisfies
func (rtSet *Set) GetEvents(dr *pb.DetectionResult) []*pb.Event {
	log.Println("GetEvents")
	var events []*pb.Event
	for _, realTimeFilter := range *rtSet {
		// check whether its an array query
		aSl, ok := realTimeFilter.query.(bson.A)
		if ok {
			if containsAll(dr.LabelNumber, aSl) {
				events = append(events, NewEvent(realTimeFilter.eventFilter, *dr))
			}
		} else {
			// check whether its a logicalQuery
			f := realTimeFilter.query.(logicalQuery)
			m, ok := f[OrOp]
			if ok {
				if compareOr(dr.LabelNumber, m) {
					events = append(events, NewEvent(realTimeFilter.eventFilter, *dr))
				}
			} else {
				// if not "or", then its "and"
				m = f[AndOp]
				if compareAnd(dr.LabelNumber, m) {
					events = append(events, NewEvent(realTimeFilter.eventFilter, *dr))
				}
			}
		}
	}
	return events
}

func (rtSet *Set) String() string {
	ret := "\n"
	for _, e := range *rtSet {
		ret += fmt.Sprintf("e.eventFilters: %v\ne.query: [%v],\n", e.eventFilter, e.query)
	}
	return ret
}
