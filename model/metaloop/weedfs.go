package metaloop

import (
	"github.com/minio/minio-go"
)

type S3ConfigResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		S3 string `json:"s3"`
	} `json:"data"`
}

var (
	endpoint  = "your_weedfs_endpoint"
	accessKey = "your_access_key"
	secretKey = "your_secret_key"
)

type S3ClientOps struct {
	accessKey string
	aecretKey string
	endpoint  string
	Client    *minio.Client
}

var S3Client S3ClientOps

//func (s *S3ClientOps) ConnWeedfs() error {
//	var err error
//	err = s.queryS3session()
//	if err != nil {
//		return err
//	}
//	s.Client, err = minio.New(endpoint, &minio.Options{
//		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
//		Secure: false, // Set to true if using HTTPS
//	})
//	return err
//
//}

//func (s *S3ClientOps) queryS3session() error {
//	var dataSetResp S3ConfigResp
//	var endpoint string
//
//	url := fmt.Sprintf("%s/api/v1/api_s3_storage_config?bucket=hsr", MClient.Url)
//
//	resp, err := MClient.Cli.R().Get(url)
//	if err != nil {
//		return err
//	}
//
//	err = json.Unmarshal(resp.Body(), &dataSetResp)
//	if err != nil {
//		return err
//	}
//
//	re := regexp.MustCompile(`^s3://(.*?):(.*?)@(.*?)\?sslmode\=(.*?)$`)
//	matches := re.FindStringSubmatch(dataSetResp.Data.S3)
//
//	s.accessKey = matches[1]
//	s.aecretKey = matches[2]
//	if matches[4] == "disable" {
//		endpoint = "http://" + matches[3]
//	} else {
//		endpoint = "https://" + matches[3]
//	}
//	s.endpoint = endpoint
//
//	return nil
//}

//func SaveImageToS3(contentType, outPath string, size int64, reader io.Reader) error {
//	_, err := S3Client.Client.PutObject(context.Background(), "hsr", outPath, reader, size, minio.PutObjectOptions{
//		ContentType: contentType,
//	})
//	return err
//}
