package metaloop

import "testing"

var bucket Bucket

func TestGetChildDir(t *testing.T) {
	createClient()

	versionId, err := B.GetDataSetVersionId("宁波-5号线")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(versionId)

	ChildDir, err := B.GetALLChildDir(versionId, "正常数据")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ChildDir)

	// nameList, err := ChildDir.GetChildDirForList()
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// t.Log(nameList)

}
