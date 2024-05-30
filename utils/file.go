package utils

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func GetFiles(pathName string, recursive bool, f *[]string) error {
	rd, err := ioutil.ReadDir(pathName)
	if err != nil {
		return err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			if recursive == true {
				err = GetFiles(pathName+"/"+fi.Name(), recursive, f)
				if err != nil {
					return err
				}
			}
		} else {
			*f = append(*f, pathName+"/"+fi.Name())
		}
	}
	return nil
}

func ReadExcel(excelPath string) (*[][]string, error) {
	var content [][]string
	xlsxRead, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, err
	}
	sheet := xlsxRead.GetActiveSheetIndex()
	name := xlsxRead.GetSheetName(sheet)
	// log.Println("sheet name is " + name)
	content = xlsxRead.GetRows(name)
	// log.Println(content)
	return &content, nil
}

// Zip 压缩文件
func Zip(srcFile string, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func CompressFiles(fileList []string, zipFile string) error {
	// 创建一个新的zip文件
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer zipWriter.Close()

	// 创建一个zip.Writer
	zipWriterObj := zip.NewWriter(zipWriter)
	defer zipWriterObj.Close()

	// 遍历文件列表
	for _, file := range fileList {
		err = addFileToZip(zipWriterObj, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, file string) error {
	// 打开文件
	fileToZip, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// 获取文件信息
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	// 创建一个zip文件的头部信息
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// 设置文件名（包含目录结构）
	header.Name = file

	// 创建一个zip文件的写入器
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 将文件内容写入zip文件
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return err
	}

	return nil
}

func Tar(source string) error {
	// tar --directory=files/source_data_export/20220324-121511 -cvf files/source_data_export/20220324-121511/0.tar 0/
	splitResult := strings.Split(source, "/")
	basePath := strings.Join(splitResult[0:len(splitResult)-1], "/")
	dirName := splitResult[len(splitResult)-1]

	cmd := exec.Command("tar", "--directory", basePath, "-cvf", source+".tar", dirName)

	//阻塞至等待命令执行完成
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		filePath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	if IsUrl(path) {
		code, err := HTTPClient("GET", path, nil, nil)
		if code != 200 || err != nil {
			return false
		}
	} else {
		_, err := os.Stat(path) //os.Stat获取文件信息
		if err != nil {
			if os.IsExist(err) {
				return true
			}
			return false
		}
	}
	return true
}

func Copy(srcFile, dstFile string) error {
	sourceFileStat, err := os.Stat(srcFile)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", srcFile)
	}

	source, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// CopyDirPrefixFiles copy指定目录下以prefix开头的文件到目标目录
func CopyDirPrefixFiles(srcPath, dstPath string, prefix string, recursive bool) error {
	if !Exists(dstPath) {
		if err := os.MkdirAll(dstPath, fs.ModePerm); err != nil {
			return err
		}
	}
	// 遍历源目录下的所有文件
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if recursive {
				err = CopyDirPrefixFiles(path, dstPath, prefix, recursive)
				if err != nil {
					return err
				}
			}
		} else {
			if strings.HasPrefix(info.Name(), prefix) {
				// 创建目标文件
				dstFileName := filepath.Join(dstPath, info.Name())
				err = os.MkdirAll(filepath.Dir(dstFileName), os.ModePerm)
				if err != nil {
					return err
				}
				// 打开源文件
				srcFile, err := os.Open(path)
				if err != nil {
					return err
				}

				// 打开目标文件
				dstFile, err := os.Create(dstFileName)
				if err != nil {
					return err
				}

				// 复制文件
				_, err = io.Copy(dstFile, srcFile)
				if err != nil {
					return err
				}

				// 关闭文件
				srcFile.Close()
				dstFile.Close()
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func Remove(filePath string) error {
	var err error
	if !Exists(filePath) {
		err = fmt.Errorf("路径不存在")
		return err
	}
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

// BatchRemove 批量删除文件，删除失败继续执行
func BatchRemove(filePaths []string) error {
	var errFileExists []string
	var errRemove []string
	for _, filePath := range filePaths {
		if !Exists(filePath) {
			errFileExists = append(errFileExists, filePath)
			continue
		}
		err := os.RemoveAll(filePath)
		if err != nil {
			errRemove = append(errRemove, filePath)
		}
	}
	if len(errFileExists) != 0 || len(errRemove) != 0 {
		return fmt.Errorf("删除文件失败 路径不存在[%v] 删除失败[%v]", errFileExists, errRemove)
	}
	return nil
}

func WriteJson(data interface{}, path string) error {
	// fmt.Println(data)
	// jsonByte, err := json.Marshal(data)
	jsonByte, err := json.MarshalIndent(data, "", "	")
	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.Write(jsonByte)
	if err != nil {
		return err
	}
	return nil
}

func ReadJson(path string) (interface{}, error) {
	//打开文件
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	//读取为[]bytes类型
	byteData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var returnData interface{}
	err = json.Unmarshal(byteData, &returnData)
	if err != nil {
		return nil, err
	}

	return returnData, nil
}

// PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
