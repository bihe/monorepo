package upload

// Config defines relevant values for the upload logic
type Config struct {
	// AllowedFileTypes is a list of mime-types allowed to be uploaded
	AllowedFileTypes []string
	// MaxUploadSize defines the maximum permissible fiile-size
	MaxUploadSize int64
	// UploadPath defines a directory where uploaded files are stored
	UploadPath string
}
