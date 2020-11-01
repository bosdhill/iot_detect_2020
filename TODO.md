# Edge
- in interfaces.proto, change:
```proto
message Labels {
    map<string, bool> labels = 1;
}

```
to 
```proto
message Labels {
    map<string, google.protobuf.Empty> labels = 1;
}

```
and change `classNamesMap` to type `map[string]struct{}`. Refer to https://stackoverflow.com/a/10486196/10741562

# Client
- The upload should be a client streaming rpc -- the return value doesn't matter to the client
- Look into Intel https://software.intel.com/en-us/articles/intel-movidius-neural-compute-stick that can interface with https://github.com/dymat/GOLOv2
- Since NMSBoxes is not yet implemented in gocv, we can leave that up to the application to annotate

# Interfaces
- Split up into client - edge interface, cloud - edge interface, app - edge interface

# Docker and Kubernetes
- Eventually containerize components in client, edge, and cloud 

# Events and Actions
- Action-on-detect is a form of complex query
- Events are conditions that can trigger Actions
- Need search syntax or query engine to parse the complex queries
## What type of unique conditions can events hold?
- Proximity conditions (close, far) between bounding boxes of objects in a frame
    - e.g. person on a bike 
    - e.g. people that are close together
- Number of different objects occurring together in a frame
    - e.g. at least 1 person and 1 bus
- Number of similar objects occurring together in a frame
    - e.g. at most 5 people
- Label with above certain threshold
    - e.g. person with conf 0.5
- Combination of the above
    - e.g. a person on a bike and at least 3 people close together in a frame
- Events also hold flags specifying the data granularity
    - Such as including metadata, image frame itself, annotated frame, etc.
    
- Need way to factor in time and strict subsets for EventConditions (such as trigger an event if only certain labels and no more)
- Proximity contains
    - Like if a person has a hat, the hat object would be inside the person bounding box
- Object detects larger things, lets say a car, and the application can check other features of that car like color, so less processing is done on the application
    - Application can extract more features, OR the cloud can extract more features (off line processing)
## What does an Action hold
- The name of the Event that triggered this action (hash of the unique conditions)
- The data that corresponds to the Event, such as:
    - Image frame (jpg)
    - annotated image frame (jpg)
    - metadata
        - labels
        - bounding box frames (annotations)
        - label confidence
- The Action would be handled by the appropriate callback method defined by the application (implement callback interface)

This is clearly 

## Event query object (defined by user/application)
``` json
{
    labels: [...], (sorted)
    conf_thresholds: [...], (maps to index of labels)
    quantity: [...], (maps to index of labels)
    quantity_bounds: [...], (maps to index of labels)
    proximities: [...], (maps to index of labels)
    proximity_distance_measure: enum
    data_return_flags: enum (bitwise XOR flags?)
}
```



## Caching
- The Event "queries" can happen in real time (right after an object is detected), and can be cached on the Edge and Cloud 
until the application makes a request such as `PullActions` or `PullEvents`
- Application framework should automatically queue or cache Actions and then allow interfaces that pull from the queue, such as worker routines

 
