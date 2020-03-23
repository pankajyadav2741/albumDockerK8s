# albumDockerK8s
Simple Go web application deployed on Docker and Kubernetes using GoCQL (Cassandra DB)

Endpoints:
----------
//Show album
GET /
//Create a new album
POST /{album}
//Delete an existing album
DELETE /{album}
//Show all images in an album
GET /{album}
//Show a particular image inside an album
GET /{album}/{image}
//Create an image in an album
POST /{album}/{image}
//Delete an image in an album
DELETE /{album}/{image}
