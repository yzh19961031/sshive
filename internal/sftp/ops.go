// internal/sftp/ops.go
package sftp

import (
	"fmt"
	"io"
	"os"

	pkgsftp "github.com/pkg/sftp"
)

// listDir 列出目录内容
func listDir(client *pkgsftp.Client, path string) ([]FileInfo, error) {
	entries, err := client.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("readdir %s: %w", path, err)
	}
	files := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		files = append(files, FileInfo{
			Name:    e.Name(),
			Size:    e.Size(),
			IsDir:   e.IsDir(),
			ModTime: e.ModTime().Unix(),
			Mode:    e.Mode().String(),
		})
	}
	return files, nil
}

// uploadFile 上传文件到远程，从 reader 读取数据
func uploadFile(client *pkgsftp.Client, remotePath string, reader io.Reader) error {
	f, err := client.OpenFile(remotePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// downloadFile 下载远程文件，返回内容字节
func downloadFile(client *pkgsftp.Client, remotePath string) ([]byte, error) {
	f, err := client.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("open remote file: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// deleteFile 删除文件或空目录
func deleteFile(client *pkgsftp.Client, path string) error {
	stat, err := client.Stat(path)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	if stat.IsDir() {
		return client.RemoveDirectory(path)
	}
	return client.Remove(path)
}

// renameFile 重命名/移动文件
func renameFile(client *pkgsftp.Client, oldPath, newPath string) error {
	return client.Rename(oldPath, newPath)
}

// mkdirAll 递归创建目录
func mkdirAll(client *pkgsftp.Client, path string) error {
	return client.MkdirAll(path)
}
