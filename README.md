goface
======

An HTTP api for detecting faces in images. The server 
will receive an image and pass it along to OpenCV. Face detection
will be done and the result returned as JSON. The JSON contains
information about the face location (the rectangular area) and a
url of the image marked up with detection info.

Features such as gender, age and mood will be attempted.


