package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"

	"github.com/johannesboyne/gofakes3"
	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	libhttp "github.com/rclone/rclone/lib/http"
	"github.com/rclone/rclone/vfs"
	"github.com/rclone/rclone/vfs/vfsflags"
	"github.com/spf13/cobra"
)

func init() {
	flagSet := Command.Flags()
	vfsflags.AddFlags(flagSet)
}

var Command = &cobra.Command{
	Use:   "s3 remote:path",
	Short: "Serve the remote over s3.",
	Long:  "Run a s3 server to serve a remote",
	RunE: func(command *cobra.Command, args []string) error {
		cmd.CheckArgs(1, 1, command, args)
		f := cmd.NewFsSrc(args)

		cmd.Run(false, false, command, func() error {
			s, err := newS3(context.Background(), f)
			if err != nil {
				return err
			}

			s.Serve()
			s.Wait()
			return nil
		})
		return nil
	},
}

type S3 struct {
	*libhttp.Server
	f       fs.Fs
	_vfs    *vfs.VFS
	handler http.Handler
	ctx     context.Context
}

type S3Backend struct {
	fs *vfs.VFS
	s3 *S3
}

func newS3(ctx context.Context, f fs.Fs) (s3 *S3, err error) {
	s3 = &S3{
		f:    f,
		ctx:  ctx,
		_vfs: vfs.New(f, &vfsflags.Opt),
	}

	s3faker := gofakes3.New(&S3Backend{
		fs: s3._vfs,
		s3: s3,
	})

	s3.handler = s3faker.Server()

	s3.Server, err = libhttp.NewServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init server: %w", err)
	}

	router := s3.Server.Router()
	router.Use(AuthMiddleware)
	router.Handle("/*", s3.handler)

	return s3, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform custom operations before passing the request to the next handler

		headers := r.Header
		token := headers["Authorization"][0]
		re := regexp.MustCompile(`Credential=[^\/]+`)
		token = re.FindString(token)
		token = token[11:]
		fmt.Println(token)
		// fmt.Println(r.Header)
		// fmt.Println(r.URL)
		// fmt.Println(r.Method)
		// AWS4-HMAC-SHA256 Credential=<oauth-access-token>/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-date,Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024
		r.Header.Set("Authorization", "Bearer "+token)
		// Call the next handler
		next.ServeHTTP(w, r)

		// Perform custom operations after the request has been handled by other handlers
		// TODO check headers here
		fmt.Println(r.Header)
	})
}

func (backend *S3Backend) ListBuckets() ([]gofakes3.BucketInfo, error) {
	// using fs directly
	// items, err := backend.s3.f.List(backend.s3.ctx, "")
	// fmt.Println(items)
	// for _, item := range items {
	// 	if a, ok := item.(fs.Directory); ok {
	// 		fmt.Println(ok)
	// 		fmt.Println(a)
	// 	}
	// }
	node, err := backend.fs.Stat("/")
	if err != nil {
		return nil, err
	}

	if !node.IsDir() {
		return nil, fmt.Errorf("No such key: /")
	}

	dir := node.(*vfs.Dir)
	entries, err := dir.ReadDirAll()
	if err != nil {
		return nil, err
	}

	var response []gofakes3.BucketInfo
	for _, entry := range entries {
		if entry.IsDir() {
			response = append(response, gofakes3.BucketInfo{
				Name:         url.QueryEscape(entry.Name()),
				CreationDate: gofakes3.NewContentTime(entry.ModTime()),
			})
		}
	}
	return response, nil
}

func (backend *S3Backend) BucketExists(bucket string) (bool, error) {
	node, err := backend.fs.Stat(bucket)
	if err != nil {
		return false, err
	}

	return node.IsDir(), nil
}

// TODO: implement
func (backend *S3Backend) ListBucket(name string, prefix *gofakes3.Prefix, page gofakes3.ListBucketPage) (*gofakes3.ObjectList, error) {
	return nil, nil
}

func (backend *S3Backend) CreateBucket(name string) error {
	return nil
}

func (backend *S3Backend) DeleteBucket(name string) error {
	return nil
}

func (backend *S3Backend) GetObject(bucketName, objectName string, rangeRequest *gofakes3.ObjectRangeRequest) (*gofakes3.Object, error) {
	return nil, nil
}

func (backend *S3Backend) HeadObject(bucketName, objectName string) (*gofakes3.Object, error) {
	return nil, nil
}

func (backend *S3Backend) DeleteObject(bucketName, objectName string) (gofakes3.ObjectDeleteResult, error) {
	result := gofakes3.ObjectDeleteResult{}
	return result, nil
}

func (backend *S3Backend) PutObject(bucketName, key string, meta map[string]string, input io.Reader, size int64) (gofakes3.PutObjectResult, error) {
	result := gofakes3.PutObjectResult{}
	return result, nil
}

func (backend *S3Backend) DeleteMulti(bucketName string, objects ...string) (gofakes3.MultiDeleteResult, error) {
	result := gofakes3.MultiDeleteResult{}
	return result, nil
}

func logCallStack() {
	pc := make([]uintptr, 10) // Adjust the buffer size as needed
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		fmt.Println("Function:", frame.Function)
		if !more {
			break
		}
	}
}
