# 图片上传和预览服务

上传api:

`http --multipart POST localhost:9000/upload bucket=b1 file_name@/Downloads/1.jpeg`

预览api:

`http://localhost:9000/preview/b1/06f53cce7fec99ce242a85275285a2d3.jpeg`

格式: {base_url}/{bucket}/{image_name}
# image_upload_server
# image_upload_server
