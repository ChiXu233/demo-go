package metaloop

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Bucket struct{}

var B Bucket

type DataSetResp struct {
	Code int `json:"code"`
	Data []struct {
		Versions []struct {
			Id      string `json:"id"`
			Comment string `json:"comment"`
		} `json:"versions"`
	} `json:"data"`
}

type ObjectDirResp struct {
	Code int            `json:"code"`
	Data ChildObjectDir `json:"data"`
	Msg  string         `json:"msg"`
}

type ObjectFileResp struct {
	Code       int               `json:"code"`
	Data       []ChildObjectFile `json:"data"`
	Msg        string            `json:"msg"`
	TotalCount int64             `json:"total_count"`
	Count      int64             `json:"count"`
}

type ChildObjectFile struct {
	Path   string `json:"path"`
	ObjUrl string `json:"obj_url"`
	Name   string `json:"name"`
}

type ChildObjectDir struct {
	ChildObjectDir []ChildObjectDir `json:"child_object_dir"`
	Name           string           `json:"name"`
	Id             string           `json:"id"`
}

// type DiskItem struct {
// 	Name       string `json:"name"`
// 	Last       bool   `json:"last"`
// 	DataSource string `json:"data_source"`
// }

func (c *ChildObjectDir) DirNameList() []string {
	var nameList []string

	for _, childObjectDir := range c.ChildObjectDir {
		nameList = append(nameList, childObjectDir.Name)
	}
	return nameList
}

func (c *ChildObjectDir) ImageList() []string {
	var imageList []string

	for _, childObjectDir := range c.ChildObjectDir {
		imageList = append(imageList, childObjectDir.Name)
	}
	return imageList
}

func (b Bucket) GetALLChildDir(datasetId, basePath string) (*ChildObjectDir, error) {
	var err error
	var objectDirResp ObjectDirResp

	url := fmt.Sprintf("%s/api/v1/search/dataset/%s/object_dir", MClient.Url, datasetId)

	resp, err := MClient.Cli.R().SetBody(map[string]string{
		"dataset_id": datasetId,
	}).Post(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &objectDirResp)
	if err != nil {
		return nil, err
	}

	checkObjectDir := objectDirResp.Data
	if strings.Contains(basePath, "/") {
		pathList := strings.Split(basePath, "/")
		for _, path := range pathList[1:] {
			if path == "" {
				continue
			}
			checkObjectDir.checkObjectDir(path)
		}
	}
	return &checkObjectDir, nil
}

func (c *ChildObjectDir) checkObjectDir(path string) {
	for _, childObjectDir := range c.ChildObjectDir {
		if childObjectDir.Name == path {
			*c = childObjectDir
		}
	}
}

func (b Bucket) GetDataSetObjectDirId(datasetId, path string) (*ChildObjectDir, error) {
	childObjectDir, err := b.GetALLChildDir(datasetId, path)
	if err != nil {
		return nil, err
	}
	return childObjectDir, err
}

func (b Bucket) GetDataSetVersion() (*DataSetResp, error) {
	var dataSetResp DataSetResp

	url := fmt.Sprintf("%s/api/v1/search/dataset", MClient.Url)

	resp, err := MClient.Cli.R().SetBody(map[string]interface{}{
		"semantic":           "轨交运维案场原始数据",
		"order_by":           "cts",
		"limit":              10,
		"need_import_status": true,
		"offset":             0,
	}).Post(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &dataSetResp)
	if err != nil {
		return nil, err
	}

	return &dataSetResp, nil
}

func (b Bucket) GetDataSetVersionId(versionName string) (string, error) {
	dataSetResp, err := b.GetDataSetVersion()
	if err != nil {
		return "", err
	}

	for _, version := range dataSetResp.Data[0].Versions {
		if version.Comment == versionName {
			return version.Id, nil
		}
	}

	return "", nil
}

func (b Bucket) GetFilesForMetaloop(path, datasetId string, recursive bool, files *[]string) error {
	objectDir, err := b.GetDataSetObjectDirId(datasetId, path)
	if err != nil {
		return err
	}
	return b.GetALLChildFile(datasetId, objectDir, files)

	// childObjectDirs, err := b.GetALLChildDir(datasetId, path)
	// if err != nil {
	// 	return err
	// }
	// for _, childObjectDir := range childObjectDirs.ChildObjectDir {
	// 	return b.GetFilesForMetaloop(fmt.Sprintf("%s/%s", path, childObjectDir.Name), datasetId, recursive, files)
	// }

}

func (b Bucket) GetALLChildFile(datasetId string, objectDir *ChildObjectDir, files *[]string) error {
	var err error
	var objectFileResp ObjectFileResp
	var offset int64
	var objectDirIds []string
	url := fmt.Sprintf("%s/api/v1/search/dataset/%s/object", MClient.Url, datasetId)

	b.getObjectDirIds(objectDir, &objectDirIds)

QUERY:
	resp, err := MClient.Cli.R().SetBody(map[string]interface{}{
		"dataset_id":        datasetId,
		"object_dir_ids":    objectDirIds,
		"limit":             40,
		"offset":            offset,
		"recursion_objects": true,
	}).Post(url)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp.Body(), &objectFileResp)
	if err != nil {
		return err
	}

	for _, childObjectFile := range objectFileResp.Data {
		*files = append(*files, childObjectFile.ObjUrl)
	}

	if objectFileResp.TotalCount > offset+40 {
		offset = offset + 40
		goto QUERY
	}
	return nil
}

func (b Bucket) getObjectDirIds(objectDir *ChildObjectDir, objectDirIds *[]string) {
	*objectDirIds = append(*objectDirIds, objectDir.Id)
	for _, object := range objectDir.ChildObjectDir {
		b.getObjectDirIds(&object, objectDirIds)
	}
}
