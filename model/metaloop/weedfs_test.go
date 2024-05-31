package metaloop

import "testing"

func TestConnWeedfs(t *testing.T) {

	createClient()

	err := S3Client.queryS3session()
	if err != nil {
		t.Error(err)
	}
	t.Logf("accessKey: %s,  aecretKey: %s, endpoint: %s \n", S3Client.accessKey, S3Client.aecretKey, S3Client.endpoint)
}
