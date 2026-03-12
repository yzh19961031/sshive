package sftp

// MsgType 消息类型
type MsgType string

const (
	// 客户端 → 服务端
	MsgList     MsgType = "list"
	MsgUpload   MsgType = "upload"
	MsgChunk    MsgType = "chunk"
	MsgDownload MsgType = "download"
	MsgDelete   MsgType = "delete"
	MsgRename   MsgType = "rename"
	MsgMkdir    MsgType = "mkdir"

	// 服务端 → 客户端
	MsgResult   MsgType = "result"
	MsgFileData MsgType = "file_data"
	MsgError    MsgType = "error"
)

// Request 客户端发来的命令
type Request struct {
	Type    MsgType `json:"type"`
	Path    string  `json:"path,omitempty"`
	NewPath string  `json:"new_path,omitempty"`
	Name    string  `json:"name,omitempty"`
	Size    int64   `json:"size,omitempty"`
	ReqID   string  `json:"req_id,omitempty"`
}

// FileInfo 文件/目录信息
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime int64  `json:"mod_time"`
	Mode    string `json:"mode"`
}

// Response 服务端返回的结果
type Response struct {
	Type    MsgType    `json:"type"`
	ReqID   string     `json:"req_id,omitempty"`
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
	Name    string     `json:"name,omitempty"`
	Size    int64      `json:"size,omitempty"`
}
