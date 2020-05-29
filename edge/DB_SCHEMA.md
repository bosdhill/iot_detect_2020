# Tables

The object detection results will be stored in 3 tables, all identified by a primary key which is the `detection time`
- image_table
The image table stores the image row, which consists of  `detection_time` (primary key), the `go.Mat` image, and `empty` which is used to determine if there are any `labels` for this image record
- bound_box_table
The bounding box table stores the bounding box row, which consists of  `detection_time` (foreign key), the `label`, box `dimensions`, and the confidence `conf`. Note that there can be multiple rows referencing the same `detection_time` key, since there could be multiple `labels` detected on one image. 
- label_existence_table
The label existence table stores `detection_time` (foreign key), and the rest of the columns are the labels of the dataset (there are 80) which are either true or false